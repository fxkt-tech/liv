package liv

import (
	"context"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
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
			output.Map(filter.SelectStream(0, filter.StreamAudio, false).Name(0)),
			output.VideoCodec(sub.Filters.Video.Codec),
			output.AudioCodec(sub.Filters.Audio.Codec),
			output.MovFlags("faststart"),
			output.Thread(sub.Threads),
			output.MaxMuxingQueueSize(4086),
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
