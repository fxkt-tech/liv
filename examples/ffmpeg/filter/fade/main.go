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

		fBase = filter.Fade("in", 0, 1.5, "black").Use(iMain.V())
		fOpt  = filter.Fade("out", 3, 1.5, "black", filter.WithKV("alpha", 1)).Use(iMain.V())

		output = output.New(
			output.Map(fOpt),
			output.VideoCodec(codec.X264),
			output.File("out_fade.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iMain).
		AddFilter(fBase, fOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
