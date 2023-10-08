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
		params = &liv.ExtractAudioParams{
			Infile:  "../../../testdata/in.mp4",
			Outfile: "out-extract-audio.aac",
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			ffmpeg.Debug(true),
			ffmpeg.Dry(true),
		),
	)
	err := tc.ExtractAudio(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
