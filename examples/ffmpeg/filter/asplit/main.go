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

		iA1 = input.WithSimple("a1.aac")

		fBase = filter.ASplit(2).Use(iA1.A())
		fMix  = filter.AMix(2).Use(fBase.S(0), fBase.S(1))

		fOpt    = filter.ASplit(2, filter.WithKV("dummy", 1)).Use(iA1.A())
		fMixOpt = filter.AMix(2).Use(fOpt.S(0), fOpt.S(1))

		output = output.New(
			output.Map(fMixOpt),
			output.AudioCodec(codec.AAC),
			output.File("out_asplit.aac"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iA1).
		AddFilter(fBase, fMix, fOpt, fMixOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
