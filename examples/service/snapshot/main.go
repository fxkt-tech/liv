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
			StartTime: 0,
			FrameType: 2,
			Frames:    []int32{1, 10, 300, 301},
			// Num: 4,
			// Interval: 5,
			Width:  960,
			Height: 540,
		}
	)

	tc := liv.NewSnapshot(
		liv.FFmpegOptions(
			ffmpeg.WithBin("ffmpeg"),
			ffmpeg.WithDebug(true),
		),
	)
	err := tc.Simple(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
