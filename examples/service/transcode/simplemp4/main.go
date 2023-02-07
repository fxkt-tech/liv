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
			Infile: "../../../testdata/in.mp4",
			Subs: []*liv.SubTranscodeParams{
				{
					Outfile: "out-simple-mp4.mp4",
					Filters: &liv.Filters{
						Container: codec.MP4,
						Video: &liv.Video{
							Codec:  codec.X264,
							Height: 540,
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
						Clip: &liv.Clip{
							Seek:     5,
							Duration: 10,
						},
					},
				},
				// {
				// 	Outfile: "out-simple-mp4-2.mp4",
				// 	Filters: &liv.Filters{
				// 		Video: &liv.Video{
				// 			Codec:  codec.X264,
				// 			Height: 720,
				// 		},
				// 		Audio: &liv.Audio{
				// 			Codec: codec.AAC,
				// 		},
				// 	},
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
	err := tc.SimpleMP4(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
