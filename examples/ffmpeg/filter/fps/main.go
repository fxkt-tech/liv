package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	livmath "github.com/fxkt-tech/liv/pkg/math"
)

func main() {
	var (
		ctx = context.Background()

		iMain = input.WithSimple("in.mp4")
		fps   = &livmath.Rational[int, int]{Num: 30, Den: 1}

		fBase = filter.FPS(fps).Use(iMain.V())
		fOpt  = filter.FPS(fps, filter.WithKV("round", "up")).Use(iMain.V())

		output = output.New(
			output.Map(fOpt),
			output.VideoCodec(codec.X264),
			output.File("out_fps.mp4"),
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
