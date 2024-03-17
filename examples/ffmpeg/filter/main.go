package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

func main() {
	var (
		ctx = context.Background()

		// inputs
		iMain = input.WithSimple("in.mp4")

		// filters
		fSplit   = filter.Split(2).Use(stream.V(0))
		fOverlay = filter.Logo(50, 100, filter.LogoTopLeft).Use(fSplit.Get(0), fSplit.Get(1))
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).AddInput(
		iMain,
	).AddFilter(
		fSplit, fOverlay,
	).AddOutput(
		output.New(
			output.Map(fOverlay),
			output.Map(stream.Select(0, stream.MayAudio)),
			output.Metadata("comment", "xx"),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out.mp4"),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
