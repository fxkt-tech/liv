package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

func main() {
	var (
		ctx = context.Background()

		input1 = input.WithSimple("in2.mp4")

		scale1 = filter.Scale(-2, -2).Use(stream.V(0))

		outfolder = "outputs/"

		mainFile = outfolder + "index.m3u8"
		segFile  = outfolder + "seg%5d.ts"
	)

	err := ffmpeg.New(
		ffmpeg.WithLogLevel(ffmpeg.LogLevelError),
		ffmpeg.WithDebug(true),
		// ffmpeg.Dry(true),
	).AddInput(
		input1,
	).AddFilter(
		scale1,
	).AddOutput(
		output.New(
			output.Map(scale1),
			output.Map(stream.Select(0, stream.MayAudio)),
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
