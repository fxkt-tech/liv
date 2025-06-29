package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

func main() {
	var (
		ctx    = context.Background()
		input1 = input.WithSimple("file.txt")
	)

	err := ffmpeg.New(
		ffmpeg.WithLogLevel(""),
		ffmpeg.WithDebug(true),
		ffmpeg.WithDry(true),
	).AddInput(
		input1,
	).AddOutput(
		output.New(
			output.VideoCodec(codec.Copy),
			output.AudioCodec(codec.Copy),
			output.Format("concat"),
			output.File("out.mp4"),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
