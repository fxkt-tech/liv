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

		fBase  = filter.Split(2).Use(iMain.V())
		fMerge = filter.Overlay(0, 0).Use(fBase.S(0), fBase.S(1))

		fOpt      = filter.Split(2, filter.WithKV("dummy", 1)).Use(iMain.V())
		fMergeOpt = filter.Overlay(0, 0).Use(fOpt.S(0), fOpt.S(1))

		output = output.New(
			output.Map(fMergeOpt),
			output.VideoCodec(codec.X264),
			output.File("out_split.mp4"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iMain).
		AddFilter(fBase, fMerge, fOpt, fMergeOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
