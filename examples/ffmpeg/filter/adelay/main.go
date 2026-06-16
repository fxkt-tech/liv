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

		fBase = filter.ADelay(1.5).Use(iA1.A())
		fOpt  = filter.ADelay(1.5, filter.WithAll(0)).Use(iA1.A())

		output = output.New(
			output.Map(fOpt),
			output.AudioCodec(codec.AAC),
			output.File("out_adelay.aac"),
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
