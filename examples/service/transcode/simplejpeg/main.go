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
			Infile: "../../../testdata/in.jpg",
			Subs: []*liv.SubTranscodeParams{
				{
					Outfile: "out-simple-jpg.jpg",
					Filters: &liv.Filters{
						Container: codec.JPG,
						Video: &liv.Video{
							Codec:  codec.MJPEG,
							Width:  -2,
							Height: 540,
						},
						Delogo: []*liv.Delogo{
							{
								Rect: &liv.Rectangle{
									X: 10, Y: 10, W: 200, H: 50,
								},
							},
						},
						Logo: []*liv.Logo{
							{
								File: "../../../testdata/logo.png",
								Pos:  "TopRight",
								Dx:   10,
								Dy:   10,
							},
						},
					},
				},
			},
		}
	)

	tc := liv.NewTranscode(
		liv.FFmpegOptions(
			ffmpeg.WithBin("ffmpeg"),
			ffmpeg.WithDebug(true),
			// ffmpeg.Dry(true),
		),
	)
	err := tc.SimpleJPEG(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
