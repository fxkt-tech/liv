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
		params = &liv.SnapshotParams{
			Infile:    "../../testdata/in.mp4",
			Outfile:   "ss/simple-%05d.jpg",
			StartTime: 3,
			FrameType: 1,
			Num:       1,
			// Interval:  1,
			Width:  960,
			Height: 540,
		}
	)

	tc := liv.NewSnapshot(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			ffmpeg.Debug(true),
		),
	)
	err := tc.Simple(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
