package liv

import (
	"context"

	"fxkt.tech/liv/ffmpeg"
	"fxkt.tech/liv/ffmpeg/codec"
	"fxkt.tech/liv/ffmpeg/filter"
	"fxkt.tech/liv/ffmpeg/input"
	"fxkt.tech/liv/ffmpeg/naming"
	"fxkt.tech/liv/ffmpeg/output"
)

type TranscodeOption func(*Transcode)

func FFmpegOptions(ffmpegOpts ...ffmpeg.FFmpegOption) TranscodeOption {
	return func(t *Transcode) {
		t.ffmpegOpts = ffmpegOpts
	}
}

func FFprobeOptions(ffprobeOpts ...ffmpeg.FFprobeOption) TranscodeOption {
	return func(t *Transcode) {
		t.ffprobeOpts = ffprobeOpts
	}
}

type Transcode struct {
	ffmpegOpts  []ffmpeg.FFmpegOption
	ffprobeOpts []ffmpeg.FFprobeOption

	spec *TranscodeSpec
}

func NewTranscode(opts ...TranscodeOption) *Transcode {
	tc := &Transcode{
		spec: NewTranscodeSpec(),
	}
	for _, opt := range opts {
		opt(tc)
	}
	return tc
}

func (tc *Transcode) SimpleMP4(ctx context.Context, params *TranscodeParams) error {
	err := tc.spec.CheckSatified(params)
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
		var lastFilter filter.Filter

		// 处理遮标
		if delogos := sub.Filters.Delogo; len(delogos) > 0 {
			for _, delogo := range delogos {
				fdelogo := filter.Delogo(nm.Gen(), int64(delogo.Rect.X), int64(delogo.Rect.Y), int64(delogo.Rect.W), int64(delogo.Rect.H))
				filters = append(filters, fdelogo)
				lastFilter = fdelogo
			}
		}

		// 视频缩放
		if sub.Filters.Video != nil {
			scale := filter.Scale(nm.Gen(), sub.Filters.Video.Width, sub.Filters.Video.Height).Use(fsplit.Copy(i))
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
			output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.AAC),
			output.File(sub.Outfile),
		}
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
