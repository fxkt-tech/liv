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
		params = &liv.TranscodeParams{
			Infile: "in1.mp4",
			Subs: []*liv.SubTranscodeParams{
				{
					Outfile: "out1.mp4",
					Filters: &liv.Filters{
						Container: codec.MP4,
						Video: &liv.Video{
							Height: 540,
						},
						Logo: []*liv.Logo{
							{
								File: "logo1.png",
								Dx:   10,
								Dy:   8,
								Pos:  "TopRight",
							},
						},
						Clip: &liv.Clip{
							Seek:     5,
							Duration: 10,
						},
					},
				},
				{
					Outfile: "out2.mp4",
					Filters: &liv.Filters{
						Video: &liv.Video{
							Height: 720,
						},
					},
				},
			},
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.Binary("ffmpeg"),
			// ffmpeg.Dry(true),
		),
	)
	err := tc.SimpleMP4(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
