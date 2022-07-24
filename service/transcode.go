package echosrv

import (
	"context"
	"errors"
	"fmt"

	echotool "fxkt.tech/echo/tool"
	"fxkt.tech/echo/tool/codec"
	"fxkt.tech/echo/tool/filter"
	"fxkt.tech/echo/tool/input"
	"fxkt.tech/echo/tool/output"
)

var (
	ErrParamsInvalid = errors.New("params is invalid")
)

func SimpleTranscodeMP4(ctx context.Context, params *TranscodeParams, options ...echotool.FFmpegOption) error {
	if params == nil || len(params.Subs) == 0 {
		return ErrParamsInvalid
	}

	var (
		inputs   input.Inputs
		filters  filter.Filters
		outputs  output.Outputs
		sublen   = len(params.Subs)
		firstSub = params.Subs[0]
	)

	// 处理input
	//
	// 处理clip对input参数的影响
	if firstSub.Filters.Clip != nil {
		inputs = append(inputs, input.WithTime(firstSub.Filters.Clip.Duration, firstSub.Filters.Clip.Duration, params.Infile))
	} else {
		inputs = append(inputs, input.WithSimple(params.Infile))
	}
	// 处理logo对inputs数量的影响
	if firstSub.Filters.Logo != nil {
		for _, logo := range firstSub.Filters.Logo {
			inputs = append(inputs, input.WithSimple(logo.File))
		}
	}

	// 处理filter和output
	fsplit := filter.Split("fsplit", sublen)
	filters = append(filters, fsplit)
	for i, sub := range params.Subs {
		// 处理filter
		var lastfilter filter.Filter
		scale := filter.Scale(fmt.Sprintf("fscale%d", i), sub.Filters.Video.Width, sub.Filters.Video.Height).Use(fsplit.Copy(i))
		filters = append(filters, scale)
		lastfilter = scale
		if logos := sub.Filters.Logo; len(logos) > 0 {
			for i, logo := range logos {
				flogo := filter.Logo(fmt.Sprintf("flogo%d", i), int64(logo.Dx), int64(logo.Dy), filter.LogoPos(logo.Pos)).Use(lastfilter)
				filters = append(filters, flogo)
				lastfilter = flogo
			}
		}

		// 处理output
		outputs = append(outputs, output.New(
			output.Map(lastfilter.Name(0)),
			output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.AAC),
			output.File(sub.Outfile),
		))
	}

	echotool.NewFFmpeg(options...).
		AddInput(inputs...).
		AddFilter(filters...).
		AddOutput(outputs...).
		DryRun()
	return nil
	// return echotool.NewFFmpeg(options...).
	// 	AddInput(inputs...).
	// 	AddFilter(filters...).
	// 	AddOutput(outputs...).
	// 	Run(ctx)
}
