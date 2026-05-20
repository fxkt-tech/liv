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

		fBase = filter.LoudnormDoublePass(-16, 11, -1.5, -20, 8, -2, -31).Use(iA1.A())
		fOpt  = filter.LoudnormDoublePass(-16, 11, -1.5, -20, 8, -2, -31, filter.WithKV("linear", "true")).Use(iA1.A())

		output = output.New(
			output.Map(fOpt),
			output.AudioCodec(codec.AAC),
			output.File("out_loudnorm_double_pass.aac"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddInput(iA1).
		AddFilter(fBase, fOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
