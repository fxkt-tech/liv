package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

func main() {
	var (
		ctx = context.Background()

		fBase = filter.ANullSrc(5)
		fOpt  = filter.ANullSrc(5, filter.WithSampleRate(48000))

		output = output.New(
			output.Map(fOpt),
			output.AudioCodec(codec.AAC),
			output.File("out_anullsrc.aac"),
		)
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).
		AddFilter(fBase, fOpt).
		AddOutput(output).
		Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
