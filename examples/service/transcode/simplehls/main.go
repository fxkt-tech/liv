package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv"
	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
)

func main() {
	var (
		ctx    = context.Background()
		params = &liv.TranscodeSimpleHLSParams{
			Infile:  "../../../testdata/in.mp4",
			Outfile: "out-simple-hls.m3u8",
			Filters: &liv.Filters{
				Container: codec.MP4,
				HLS: &liv.HLS{
					HLSTime:            10,
					HLSSegmentFilename: "%05d.ts",
				},
				Video: &liv.Video{
					Codec:  codec.X264,
					Height: 540,
					Crf:    28,
					GOP:    300,
				},
				Audio: &liv.Audio{
					Codec: codec.AAC,
				},
				Logo: []*liv.Logo{
					{
						File: "../../../testdata/logo.png",
						Dx:   10,
						Dy:   8,
						Pos:  "TopRight",
						LW:   0,
						LH:   540,
					},
				},
				// Clip: &liv.Clip{
				// 	Seek:     5,
				// 	Duration: 10,
				// },
			},
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			ffmpeg.Debug(true),
			// ffmpeg.Dry(true),
		),
	)
	err := tc.SimpleHLS(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
