package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

func main() {
	var (
		ctx = context.Background()
		nm  = naming.New()

		input1 = input.WithSimple("in2.mp4")

		scale1 = filter.Scale(nm.Gen(), -2, -2).
			Use(filter.SelectStream(0, filter.StreamVideo, true))

		outfolder = "outputs/"

		mainFile = outfolder + "index.m3u8"
		segFile  = outfolder + "seg%5d.ts"
	)

	err := ffmpeg.New(
		ffmpeg.V(ffmpeg.LogLevelError),
		ffmpeg.Debug(true),
		// ffmpeg.Dry(true),
	).AddInput(
		input1,
	).AddFilter(
		scale1,
	).AddOutput(
		output.New(
			output.Map(scale1.Name(0)),
			output.Map("0:a?"),
			output.Crf(28),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.KV("sc_threshold", "0"),
			output.File(mainFile),
			output.MovFlags("faststart"),
			output.GOP(250),
			output.HLSSegmentType("mpegts"),
			output.HLSFlags("independent_segments"),
			output.HLSPlaylistType("vod"),
			output.HLSTime(10),
			output.HLSSegmentFilename(segFile),
			output.Format(codec.HLS),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
