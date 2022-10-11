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
			Infile:    "in.mp4",
			Outfile:   "ss/%05d.jpg",
			StartTime: 3,
			FrameType: 0,
			Num:       1,
			Interval:  1,
		}
	)

	tc := liv.NewSnapshot(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			// ffmpeg.Dry(true),
		),
	)
	err := tc.Simple(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
