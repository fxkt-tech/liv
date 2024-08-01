package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

func main() {
	var (
		ctx = context.Background()

		// inputs
		iMain = input.WithSimple("in.mp4")

		// filters
		fDelogo  = filter.Crop("iw*0.8", "ih*0.08", "iw*0.1", "ih*0.91").Use(iMain.V())
		fGBlur   = filter.GBlur(25).Use(fDelogo)
		fOverlay = filter.Overlay("W*0.1", "H*0.91").Use(iMain.V(), fGBlur)

		oOnly = output.New(
			output.Map(fOverlay),
			output.Map(iMain.MayA()),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iMain).
		AddFilter(fDelogo, fGBlur, fOverlay).
		AddOutput(oOnly).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
