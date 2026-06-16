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

		fBase = filter.Tile(3, 3).Use(iMain.V())
		fOpt  = filter.Tile(3, 3, filter.WithKV("overlap", 1)).Use(iMain.V())

		output = output.New(
			output.Map(fOpt),
			output.VideoCodec(codec.X264),
			output.File("out_tile.mp4"),
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
