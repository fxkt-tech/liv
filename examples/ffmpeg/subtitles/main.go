package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/filter/fsugar"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffprobe"
)

func main() {
	var (
		ctx = context.Background()

		infile = "in.mp4"

		ffp, _ = ffprobe.New(
			ffprobe.WithDebug(true),
		).Input(infile).Extract(ctx)
		vs     = ffp.GetFirstVideoStream()
		ow, oh = vs.Width, vs.Height

		// inputs
		iMain = input.WithSimple(infile)

		sw       = int32(float32(ow) * 0.8)
		sh       = int32(float32(oh) * 0.08)
		sx       = int32(float32(ow) * 0.1)
		sy       = int32(float32(oh) * 0.91)
		fontsize = sh
		marginv  = sy

		// filters
		fDelogo    = filter.Crop(sw, sh, sx, sy).Use(iMain.V())
		fGBlur     = filter.GBlur(25).Use(fDelogo)
		fOverlay   = filter.Overlay(sx, sy).Use(iMain.V(), fGBlur)
		fSubtitles = filter.Subtitles("subtitle.srt", "",
			fsugar.NewAssSubtitle().
				SetPlayResX(ow).
				SetPlayResY(oh).
				SetFontSize(fontsize).
				SetMarginV(marginv).
				// SetFontName("阿里巴巴普惠体 3.0").
				SetFontName("报隶-简").
				SetAlignment(6).
				String(),
		).Use(fOverlay)

		oOnly = output.New(
			output.Map(fSubtitles),
			output.Map(iMain.MayA()),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out2.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
		ffmpeg.WithLogLevel("info"),
	).
		AddInput(iMain).
		AddFilter(fDelogo, fGBlur, fOverlay, fSubtitles).
		AddOutput(oOnly).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
