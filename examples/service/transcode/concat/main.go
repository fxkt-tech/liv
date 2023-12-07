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
		params = &liv.ConcatParams{
			Infiles: []string{
				"../../../testdata/in1.mp4",
				"../../../testdata/in2.mp4",
			},
			ConcatFile: "mylist.txt",
			Outfile:    "out.mp4",
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			ffmpeg.Debug(true),
			// ffmpeg.Dry(true),
		),
	)
	err := tc.Concat(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
