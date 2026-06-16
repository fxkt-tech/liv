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

		iSub = input.WithSimple("sub.mp4")

		fBase = filter.Chromakey("0x00FF00", 0.10, 0.20).Use(iSub.V())
		fOpt  = filter.Chromakey("0x00FF00", 0.10, 0.20, filter.WithKV("yuv", 1)).Use(iSub.V())

		output = output.New(
			output.Map(fOpt),
			output.VideoCodec(codec.X264),
			output.File("out_chromakey.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iSub).
		AddFilter(fBase, fOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
