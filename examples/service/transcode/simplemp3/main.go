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
			Infile: "../../../testdata/in.aac",
			Subs: []*liv.SubTranscodeParams{
				{
					Outfile: "out-simple-mp3.mp3",
					Filters: &liv.Filters{
						Container: codec.MP3,
						Audio: &liv.Audio{
							Codec:   codec.MP3Lame,
							Bitrate: 64000,
						},
					},
				},
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
	err := tc.SimpleMP3(ctx, params)
	if err != nil {
		fmt.Println(err)
	}
}
