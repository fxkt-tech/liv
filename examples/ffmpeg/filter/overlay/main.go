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

		iMain = input.WithSimple("in.mp4")
		iLogo = input.WithSimple("logo.png")

		fBase = filter.Overlay("W-w-20", "H-h-20").Use(iMain.V(), iLogo.V())
		fOpt  = filter.Overlay("(W-w)/2", "(H-h)/2", filter.WithEnable("between(t,0,3)"), filter.WithShortest(1)).Use(iMain.V(), iLogo.V())

		output = output.New(
			output.Map(fOpt),
			output.Map(iMain.MayA()),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out_overlay.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iMain, iLogo).
		AddFilter(fBase, fOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
