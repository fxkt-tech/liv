package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/internal/encoding/json"
)

// ffmpeg -i voice.wav -filter_complex "loudnorm=I=-16:TP=-1:LRA=11:print_format=json" -f null -
func main() {
	var (
		ctx = context.Background()

		// inputs
		iMain = input.WithSimple("voice.wav")

		// filters
		fLoudnorm = filter.Loudnorm(-12, 7, -1.5).Use(iMain.A())
	)

	ln, err := ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
		ffmpeg.WithLogLevel(""),
	).AddInput(
		iMain,
	).AddFilter(
		fLoudnorm,
	).AddOutput(
		output.New(
			output.Map(fLoudnorm),
			output.Format("null"),
			output.File("-"),
		),
	).ExtractLoudnorm(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(json.ToString(ln))

	var (
		// filters
		fLoudnorm2 = filter.LoudnormDoublePass(
			-12, 7, -1.5,
			ln.InputI, ln.InputLRA, ln.InputTP, ln.InputThresh,
		).Use(iMain.A())

		// output
		oOnly = output.New(
			output.Map(fLoudnorm2),
			output.AudioCodec(codec.PCMS16LE),
			output.File("voice_ln.wav"),
		)
	)

	err = ffmpeg.New(
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).AddInput(
		iMain,
	).AddFilter(
		fLoudnorm2,
	).AddOutput(
		oOnly,
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
