package liv

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
	"github.com/fxkt-tech/liv/ffmpeg/util"
	"github.com/fxkt-tech/liv/internal/sugar"
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
	fsplit := filter.Split(sublen)
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理遮标
		if delogos := sub.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(
					int64(delogo.Rect.X), int64(delogo.Rect.Y),
					int64(delogo.Rect.W), int64(delogo.Rect.H),
				).Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if sub.Filters.Video != nil {
			scale := filter.Scale(
				util.FixPixelLen(sub.Filters.Video.Width),
				util.FixPixelLen(sub.Filters.Video.Height),
			).Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := sub.Filters.Logo; len(logos) > 0 {
			for _, logo := range logos {
				var finalLogoStream stream.Streamer
				logoStream := stream.V(logoStartIndex + i)
				if logo.NeedScale() {
					logoScale := filter.Scale(
						util.FixPixelLen(int32(logo.LW)),
						util.FixPixelLen(int32(logo.LH)),
					).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 处理output
		outputOpts := []output.Option{
			output.Map(lastFilter),
			output.Map(stream.Select(0, stream.MayAudio)),
			output.VideoCodec(sub.Filters.Video.Codec),
			output.AudioCodec(sub.Filters.Audio.Codec),
			output.MovFlags("faststart"),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}
		outputOpts = append(outputOpts, metadataOptionFromKV(sub.Filters.Metadata)...)

		// 处理在每一路输出流的裁剪
		if sub.Filters.Clip != nil {
			outputOpts = append(outputOpts,
				output.StartTime(sub.Filters.Clip.Seek),
				output.Duration(sub.Filters.Clip.Duration),
			)
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.New(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func metadataOptionFromKV(kvs []*KV) []output.Option {
	oos := make([]output.Option, len(kvs))
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
	fsplit := filter.ASplit(sublen).Use(stream.V(0))
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理output
		outputOpts := []output.Option{
			output.Map(lastFilter),
			output.AudioCodec(sub.Filters.Audio.Codec),
			output.AudioBitrate(sub.Filters.Audio.Bitrate),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.New(tc.ffmpegOpts...).
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
	fsplit := filter.Split(sublen)
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		lastFilter := fsplit.Copy(i)

		// 处理遮标
		if delogos := sub.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if sub.Filters.Video != nil {
			scale := filter.Scale(sub.Filters.Video.Width, sub.Filters.Video.Height).Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := sub.Filters.Logo; len(logos) > 0 {
			for _, logo := range logos {
				flogo := filter.Logo(int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).Use(lastFilter)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 处理output
		outputOpts := []output.Option{
			output.Map(lastFilter),
			output.VideoCodec(sub.Filters.Video.Codec),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
			output.File(sub.Outfile),
		}

		outputs = append(outputs, output.New(outputOpts...))
	}

	return ffmpeg.New(tc.ffmpegOpts...).
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
	outputOpts := []output.Option{
		output.VideoCodec(codec.Copy),
		output.AudioCodec(codec.Copy),
		output.MovFlags("faststart"),
		output.Thread(params.Threads),
		output.MaxMuxingQueueSize(4086),
		output.File(params.OutFile),
	}
	outputs = append(outputs, output.New(outputOpts...))

	return ffmpeg.New(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddOutput(outputs...).
		Run(ctx)
}

// 转hls
func (tc *Transcode) SimpleHLS(ctx context.Context, params *TranscodeSimpleHLSParams) error {
	err := tc.spec.SimpleHLSSatified(params)
	if err != nil {
		return err
	}

	var (
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
				fdelogo := filter.Delogo(
					int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).
					Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if params.Filters.Video != nil {
			scale := filter.Scale(
				util.FixPixelLen(params.Filters.Video.Width), util.FixPixelLen(params.Filters.Video.Height)).
				Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := params.Filters.Logo; len(logos) > 0 {
			for i, logo := range logos {
				var finalLogoStream stream.Streamer
				logoStream := stream.V(logoStartIndex + i)
				if logo.NeedScale() {
					logoScale := filter.Scale(
						util.FixPixelLen(int32(logo.LW)), util.FixPixelLen(int32(logo.LH))).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}
	} else {
		// 为使用map必须放一个filter
		scale := filter.Scale(util.FixPixelLen(0), util.FixPixelLen(0)).Use(lastFilter)
		filters = append(filters, scale)
		lastFilter = scale
	}

	// 处理output
	outputs[0] = output.New(
		output.Map(lastFilter),
		output.Map(stream.Select(0, stream.MayAudio)),
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

	return ffmpeg.New(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

// 转ts
func (tc *Transcode) SimpleTS(ctx context.Context, params *TranscodeSimpleTSParams) error {
	err := tc.spec.SimpleTSSatified(params)
	if err != nil {
		return err
	}

	var (
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
				fdelogo := filter.Delogo(
					int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H)).
					Use(lastFilter)
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if params.Filters.Video != nil {
			scale := filter.Scale(
				util.FixPixelLen(params.Filters.Video.Width), util.FixPixelLen(params.Filters.Video.Height)).
				Use(lastFilter)
			filters = append(filters, scale)
			lastFilter = scale
		}

		// 添加水印
		if logos := params.Filters.Logo; len(logos) > 0 {
			for i, logo := range logos {
				var finalLogoStream stream.Streamer
				logoStream := stream.V(logoStartIndex + i)
				if logo.NeedScale() {
					logoScale := filter.Scale(
						util.FixPixelLen(int32(logo.LW)), util.FixPixelLen(int32(logo.LH))).Use(logoStream)
					filters = append(filters, logoScale)
					finalLogoStream = logoScale
				} else {
					finalLogoStream = logoStream
				}
				flogo := filter.Logo(int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).
					Use(lastFilter, finalLogoStream)
				filters = append(filters, flogo)
				inputs = append(inputs, input.WithSimple(logo.File))
				lastFilter = flogo
			}
		}

		// 设置视频PTS
		if pts := params.Filters.Video.PTS; pts != "" {
			fsetpts := filter.SetPTS(pts).Use(lastFilter)
			filters = append(filters, fsetpts)
			lastFilter = fsetpts
		}

		// 设置音频PTS
		if apts := params.Filters.Video.APTS; apts != "" {
			aStream := stream.A(0)
			fasetpts := filter.ASetPTS(apts).Use(aStream)
			filters = append(filters, fasetpts)
			lastAudioFilter = fasetpts
		}
	} else {
		// 为使用map必须放一个filter
		scale := filter.Scale(util.FixPixelLen(0), util.FixPixelLen(0)).Use(lastFilter)
		filters = append(filters, scale)
		lastFilter = scale
	}

	// 处理output
	outputs[0] = output.New(
		output.Map(lastFilter),
		output.Map(lastAudioFilter),
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

	return ffmpeg.New(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		Run(ctx)
}

func ConcatFile(files []string, localPath string) error {
	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	files = sugar.Multi(files, func(f string) string { return fmt.Sprintf("file '%s'", f) })
	fs := strings.Join(files, "\n")
	_, err = f.Write([]byte(fs))
	return err
}

// 多个视频合并成一个
func (tc *Transcode) Concat(ctx context.Context, params *ConcatParams) error {
	err := tc.spec.ConcatSatified(params)
	if err != nil {
		return err
	}

	err = ConcatFile(params.Infiles, params.ConcatFile)
	if err != nil {
		return err
	}

	return ffmpeg.New(
		ffmpeg.WithDebug(true),
	).AddInput(
		input.WithConcat(params.ConcatFile),
	).AddOutput(
		output.New(
			output.VideoCodec(codec.Copy),
			output.AudioCodec(codec.Copy),
			output.Duration(params.Duration),
			output.File(params.Outfile),
		),
	).Run(ctx)
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
	outputOpts := []output.Option{
		output.AudioCodec(codec.Copy),
		output.VideoCodec(codec.Nope),
		output.File(params.Outfile),
	}

	outputs = append(outputs, output.New(outputOpts...))

	return ffmpeg.New(tc.ffmpegOpts...).
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
	outputOpts := []output.Option{
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

	return ffmpeg.New(tc.ffmpegOpts...).
		AddInput(inputs...).
		AddOutput(outputs...).
		Run(ctx)
}
