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

		fBase = filter.GBlur(20).Use(iMain.V())
		fOpt  = filter.GBlur(20, filter.WithSteps(3)).Use(iMain.V())

		output = output.New(
			output.Map(fOpt),
			output.VideoCodec(codec.X264),
			output.File("out_gblur.mp4"),
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
