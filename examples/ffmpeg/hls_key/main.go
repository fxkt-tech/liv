package main

import (
	"context"

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

		input1         = input.WithSimple("in.mp4")
		hlsKeyInfoFile = "hls/file.keyinfo"

		scale1 = filter.Scale(-2, -2).Use(stream.V(0))

		outfolder = "video/"

		mainFile = outfolder + "m.m3u8"
		segFile  = outfolder + "m-%5d.ts"
	)

	ffmpeg.New(
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
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File(mainFile),
			output.MovFlags("faststart"),
			output.HLSSegmentType("mpegts"),
			output.HLSFlags("independent_segments"),
			output.HLSPlaylistType("vod"),
			output.HLSTime(2),
			output.HLSKeyInfoFile(hlsKeyInfoFile), // 加密
			output.HLSSegmentFilename(segFile),
			output.Format(codec.HLS),
		),
	).Run(ctx)
}
