package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv"
	"github.com/fxkt-tech/liv/ffmpeg"
)

func main() {
	var (
		ctx    = context.Background()
		params = &liv.ConvertContainerParams{
			InFile:  "../../testdata/in.mp4",
			OutFile: "out-convert-container.mp4",
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.WithBin("ffmpeg"),
			// ffmpeg.WithDry(true),
		),
	)
	err := tc.ConvertContainer(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
