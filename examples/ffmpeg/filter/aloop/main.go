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

		fBase = filter.ALoop(-1, 0).Use(iA1.A())
		fOpt  = filter.ALoop(3, 220500, filter.WithKV("start", 0)).Use(iA1.A())

		output = output.New(
			output.Map(fOpt),
			output.AudioCodec(codec.AAC),
			output.File("out_aloop.aac"),
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
