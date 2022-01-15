package main

import (
	"context"
	"testing"

	"fxkt.tech/ffmpeg"
	"fxkt.tech/ffmpeg/codec"
	"fxkt.tech/ffmpeg/filter"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

func TestHLS(t *testing.T) {
	ff := ffmpeg.Default()
	ff.LogLevel("error")
	ff.AddInput(input.New(
		input.I("in.mp4"),
	))
	ff.AddOutput(output.New(
		output.VideoCoder(codec.VideoX264),
		output.AudioCoder(codec.Copy),
		output.Map(filter.SelectStream(0, filter.StreamVideo, true)),
		output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
		// output.Crf(25),
		output.VideoBitrate(1000000),
		output.File("m.m3u8"),
		output.MovFlags("faststart"),
		output.HlsSegmentType("mpegts"),
		output.HlsFlags("independent_segments"),
		output.HlsPlaylistType("vod"),
		output.HlsTime(2),
		output.HlsKeyInfoFile("file.keyinfo"), // 加密
		output.HlsSegmentFilename("m-%5d.ts"),
		output.Format(codec.Hls),
	))
	ff.DryRun()
	ff.Run(context.Background())
}
