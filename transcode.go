package liv

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/util"
)

type Transcode struct {
	*options

	spec *TranscodeSpec
}

func NewTranscode(opts ...Option) *Transcode {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	tc := &Transcode{
		spec:    NewTranscodeSpec(),
		options: o,
	}
	return tc
}

func (tc *Transcode) SimpleMP4(ctx context.Context, params *TranscodeParams) error {
	err := tc.spec.SimpleMP4Satified(params)
	if err != nil {
		return err
	}

	var (
		nm             = naming.New()
		inputs         input.Inputs
		filters        filter.Filters
		outputs        output.Outputs
		sublen         = len(params.Subs)
		logoStartIndex = 1
	)

	// 处理input
	inputs = append(inputs, input.WithSimple(params.Infile))

	// 处理filter和output
	//
	// 将源文件分为多个副本，用于生成多个output
	fsplit := filter.Split(nm.Gen(), sublen)
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理遮标
		if delogos := sub.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(nm.Gen(), int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if sub.Filters.Video != nil {
			scale := filter.Scale(nm.Gen(),
				util.FixPixelLen(sub.Filters.Video.Width), util.FixPixelLen(sub.Filters.Video.Height)).
				Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := sub.Filters.Logo; len(logos) > 0 {
			for _, logo := range logos {
				var finalLogoStream filter.Filter
				logoStream := filter.SelectStream(logoStartIndex+i, filter.StreamVideo, true)
				if logo.NeedScale() {
					logoScale := filter.Scale(nm.Gen(),
						util.FixPixelLen(int32(logo.LW)), util.FixPixelLen(int32(logo.LH))).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(nm.Gen(), int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 处理output
		outputOpts := []output.OutputOption{
			output.Map(lastFilter.Name(0)),
			output.Map("0:a?"),
			output.VideoCodec(sub.Filters.Video.Codec),
			output.AudioCodec(sub.Filters.Audio.Codec),
			output.MovFlags("faststart"),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}
		outputOpts = append(outputOpts, metadataOutputOptionFromKV(sub.Filters.Metadata)...)

		// 处理在每一路输出流的裁剪
		if sub.Filters.Clip != nil {
			outputOpts = append(outputOpts,
				output.StartTime(sub.Filters.Clip.Seek),
				output.Duration(sub.Filters.Clip.Duration),
			)
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func metadataOutputOptionFromKV(kvs []*KV) []output.OutputOption {
	oos := make([]output.OutputOption, len(kvs))
	for i, kv := range kvs {
		oos[i] = output.Metadata(kv.K, kv.V)
	}
	return oos
}

func (tc *Transcode) SimpleMP3(ctx context.Context, params *TranscodeParams) error {
	err := tc.spec.SimpleMP3Satified(params)
	if err != nil {
		return err
	}

	var (
		nm      = naming.New()
		inputs  input.Inputs
		filters filter.Filters
		outputs output.Outputs
		sublen  = len(params.Subs)
	)

	// 处理input
	inputs = append(inputs, input.WithSimple(params.Infile))

	// 处理filter和output
	//
	// 将源文件分为多个副本，用于生成多个output
	fsplit := filter.ASplit(nm.Gen(), sublen).Use(filter.SelectStream(0, filter.StreamAudio, true))
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理output
		outputOpts := []output.OutputOption{
			output.Map(lastFilter.Name(0)),
			output.AudioCodec(sub.Filters.Audio.Codec),
			output.AudioBitrate(sub.Filters.Audio.Bitrate),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func (tc *Transcode) SimpleJPEG(ctx context.Context, params *TranscodeParams) error {
	err := tc.spec.SimpleJPEGSatified(params)
	if err != nil {
		return err
	}

	var (
		nm      = naming.New()
		inputs  input.Inputs
		filters filter.Filters
		outputs output.Outputs
		sublen  = len(params.Subs)
	)

	// 处理input
	inputs = append(inputs, input.WithSimple(params.Infile))

	// 处理filter和output
	//
	// 将源文件分为多个副本，用于生成多个output
	fsplit := filter.Split(nm.Gen(), sublen)
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理遮标
		if delogos := sub.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(nm.Gen(), int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if sub.Filters.Video != nil {
			scale := filter.Scale(nm.Gen(), sub.Filters.Video.Width, sub.Filters.Video.Height).Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := sub.Filters.Logo; len(logos) > 0 {
			for _, logo := range logos {
				flogo := filter.Logo(nm.Gen(), int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).Use(lastFilter)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 处理output
		outputOpts := []output.OutputOption{
			output.Map(lastFilter.Name(0)),
			output.VideoCodec(sub.Filters.Video.Codec),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func (tc *Transcode) ConvertContainer(ctx context.Context, params *ConvertContainerParams) error {
	err := tc.spec.ConvertContainerSatified(params)
	if err != nil {
		return err
	}

	var (
		inputs  input.Inputs
		outputs output.Outputs
	)

	// 处理input
	inputs = append(inputs, input.WithSimple(params.InFile))

	// 处理output
	outputOpts := []output.OutputOption{
		output.VideoCodec(codec.Copy),
		output.AudioCodec(codec.Copy),
		output.MovFlags("faststart"),
		output.Thread(params.Threads),
		output.MaxMuxingQueueSize(4086),
		output.File(params.OutFile),
	}
	outputs = append(outputs, output.New(outputOpts...))

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddOutput(outputs...).
		Run(ctx)
}

func (tc *Transcode) SimpleHLS(ctx context.Context, params *TranscodeSimpleHLSParams) error {
	err := tc.spec.SimpleHLSSatified(params)
	if err != nil {
		return err
	}

	var (
		nm             = naming.New()
		inputs         input.Inputs
		filters        filter.Filters
		outputs        = make(output.Outputs, 1)
		logoStartIndex = 1 // logo输入文件的起始索引
	)

	// 处理input
	if params.Filters.Clip != nil {
		inputs = append(inputs,
			input.WithTime(params.Filters.Clip.Seek, params.Filters.Clip.Duration, params.Infile))
	} else {
		inputs = append(inputs, input.WithSimple(params.Infile))
	}

	// 处理filter和output
	//
	var lastFilter filter.Filter
	if params.Filters != nil {
		// 处理filter

		// 处理遮标
		if delogos := params.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(nm.Gen(),
					int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).
					Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if params.Filters.Video != nil {
			scale := filter.Scale(nm.Gen(),
				util.FixPixelLen(params.Filters.Video.Width), util.FixPixelLen(params.Filters.Video.Height)).
				Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := params.Filters.Logo; len(logos) > 0 {
			for i, logo := range logos {
				var finalLogoStream filter.Filter
				logoStream := filter.SelectStream(logoStartIndex+i, filter.StreamVideo, true)
				if logo.NeedScale() {
					logoScale := filter.Scale(nm.Gen(),
						util.FixPixelLen(int32(logo.LW)), util.FixPixelLen(int32(logo.LH))).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(nm.Gen(), int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}
	} else {
		// 为使用map必须放一个filter
		scale := filter.Scale(nm.Gen(), util.FixPixelLen(0), util.FixPixelLen(0)).Use(lastFilter)
		filters = append(filters, scale)
		lastFilter = scale
	}

	// 处理output
	outputs[0] = output.New(
		output.Map(lastFilter.Name(0)),
		output.Map("0:a?"),
		output.Crf(params.Filters.Video.Crf),
		output.VideoCodec(params.Filters.Video.Codec),
		output.AudioCodec(params.Filters.Audio.Codec),
		output.KV("sc_threshold", "0"),
		output.File(params.Outfile),
		output.MovFlags("faststart"),
		output.GOP(params.Filters.Video.GOP),
		output.HLSSegmentType("mpegts"),
		output.HLSFlags("independent_segments"),
		output.HLSPlaylistType("vod"),
		output.HLSTime(params.Filters.HLS.HLSTime),
		output.HLSSegmentFilename(params.Filters.HLS.HLSSegmentFilename),
		output.Format(codec.HLS),
	)

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func (tc *Transcode) SimpleTS(ctx context.Context, params *TranscodeSimpleTSParams) error {
	err := tc.spec.SimpleTSSatified(params)
	if err != nil {
		return err
	}

	var (
		nm             = naming.New()
		inputs         input.Inputs
		filters        filter.Filters
		outputs        = make(output.Outputs, 1)
		logoStartIndex = 1
	)

	// 处理input
	if params.Filters.Clip != nil {
		inputs = append(inputs, input.WithTime(params.Filters.Clip.Seek, params.Filters.Clip.Duration, params.Infile))
	} else {
		inputs = append(inputs, input.WithSimple(params.Infile))
	}

	// 处理filter和output
	//
	var lastFilter, lastAudioFilter filter.Filter
	if params.Filters != nil {
		// 处理filter

		fmt.Println("处理filter")

		// 处理遮标
		if delogos := params.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(nm.Gen(),
					int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).
					Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if params.Filters.Video != nil {
			scale := filter.Scale(nm.Gen(),
				util.FixPixelLen(params.Filters.Video.Width), util.FixPixelLen(params.Filters.Video.Height)).
				Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := params.Filters.Logo; len(logos) > 0 {
			for i, logo := range logos {
				var finalLogoStream filter.Filter
				logoStream := filter.SelectStream(logoStartIndex+i, filter.StreamVideo, true)
				if logo.NeedScale() {
					logoScale := filter.Scale(nm.Gen(),
						util.FixPixelLen(int32(logo.LW)), util.FixPixelLen(int32(logo.LH))).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(nm.Gen(), int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 设置视频PTS
		if pts := params.Filters.Video.PTS; pts != "" {
			fsetpts := filter.SetPTS(nm.Gen(), pts).Use(lastFilter)
			filters = append(filters, fsetpts)
			lastFilter = fsetpts
		}

		// 设置音频PTS
		if apts := params.Filters.Video.APTS; apts != "" {
			aStream := filter.SelectStream(0, filter.StreamAudio, true)
			fasetpts := filter.ASetPTS(nm.Gen(), apts).Use(aStream)
			filters = append(filters, fasetpts)
			lastAudioFilter = fasetpts
		}
	} else {
		// 为使用map必须放一个filter
		scale := filter.Scale(nm.Gen(), util.FixPixelLen(0), util.FixPixelLen(0)).Use(lastFilter)
		filters = append(filters, scale)
		lastFilter = scale
	}

	// 处理output
	outputs[0] = output.New(
		output.Map(lastFilter.Name(0)),
		output.Map(lastAudioFilter.Name(0)),
		output.VideoCodec(params.Filters.Video.Codec),
		output.AudioCodec(params.Filters.Audio.Codec),
		output.Crf(params.Filters.Video.Crf),
		output.KV("sc_threshold", "0"),
		output.GOP(params.Filters.Video.GOP),
		output.MovFlags("faststart"),
		output.Thread(params.Threads),
		output.MaxMuxingQueueSize(4086),
		output.File(params.Outfile),
	)

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

// ffmpeg -i in.mp4 -vn -c:a copy out.aac
func (tc *Transcode) ExtractAudio(ctx context.Context, params *ExtractAudioParams) error {
	err := tc.spec.ExtractAudioSatified(params)
	if err != nil {
		return err
	}

	var (
		inputs  input.Inputs
		outputs output.Outputs
	)

	// 处理input
	inputs = append(inputs, input.WithSimple(params.Infile))

	// 处理output
	outputOpts := []output.OutputOption{
		output.AudioCodec(codec.Copy),
		output.VideoCodec(codec.Nop),
		output.File(params.Outfile),
	}

	outputs = append(outputs, output.New(outputOpts...))

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddOutput(outputs...).
		Run(ctx)
}

// ffmpeg -i in.mp4 -vn -c:a copy out.aac
func (tc *Transcode) MergeByFrames(ctx context.Context, params *MergeParams) error {
	err := tc.spec.MergeByFramesSatified(params)
	if err != nil {
		return err
	}

	var (
		inputs  input.Inputs
		outputs output.Outputs
	)

	// 处理input
	inputs = append(inputs,
		input.New(input.FPS(params.Filters.Video.FPS), input.I(params.FramesInfile)),
		input.WithSimple(params.AudioInfile),
	)

	// 处理output
	outputOpts := []output.OutputOption{
		output.AudioCodec(codec.Copy),
		output.VideoCodec(params.Filters.Video.Codec),
		output.AudioCodec(params.Filters.Audio.Codec),
		output.PixFmt(params.Filters.Video.PixFmt),
		output.Crf(params.Filters.Video.Crf),
		output.MovFlags("faststart"),
		output.MaxMuxingQueueSize(4086),
		output.File(params.Outfile),
	}

	outputs = append(outputs, output.New(outputOpts...))

	return ffmpeg.NewFFmpeg(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddOutput(outputs...).
		Run(ctx)
}
