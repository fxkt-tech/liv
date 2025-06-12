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
		ctx    = context.Background()
		input1 = input.WithSimple("a1.aac")
		input2 = input.WithSimple("a2.aac")
		fAmix  = filter.AMix(2)
	)

	err := ffmpeg.New(
		ffmpeg.WithLogLevel(""),
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).AddInput(
		input1, input2,
	).AddFilter(
		fAmix,
	).
		AddOutput(
			output.New(
				output.Map(fAmix),
				output.AudioCodec(codec.AAC),
				output.File("out.mp4"),
			),
		).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
