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

		// inputs
		iMain = input.WithSimple("in.mp4")
		// iSub  = input.WithSimple("xx.mp4")

		// filters
		// fSplit   = filter.Split(2).Use(iMain.V(), iSub.V())
		// fOverlay = filter.Overlay(fsugar.LogoPos(50, 100, fsugar.LogoPosTopLeft)).Use(fSplit.S(0), fSplit.S(1))
		fDelogo = filter.Delogo(0, 0, 400, 400).Use(iMain.V())
	)

	err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).AddInput(
		// iMain, iSub,
		iMain,
	).AddFilter(
		// fSplit, fOverlay,
		fDelogo,
	).AddOutput(
		output.New(
			output.Map(fDelogo),
			output.Map(iMain.MayA()),
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
