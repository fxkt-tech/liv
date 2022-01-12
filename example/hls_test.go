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
		input.SetI("in.mp4"),
	))
	ff.AddOutput(output.New(
		output.SetVideoCoder(codec.VideoX264),
		output.SetAudioCoder(codec.Copy),
		output.SetMap(filter.SelectStream(0, filter.StreamVideo, true)),
		output.SetMap(filter.SelectStream(0, filter.StreamAudio, false)),
		output.SetFile("index.m3u8"),
		output.SetMovFlags("faststart"),
		output.SetHlsSegmentType("mpegts"),
		output.SetHlsFlags("independent_segments"),
		output.SetHlsPlaylistType("vod"),
		output.SetHlsTime(2),
		output.SetHlsKeyInfoFile("file.keyinfo"), // 加密
		output.SetHlsSegmentFilename("m-%5d.ts"),
		output.SetFormat(codec.Hls),
	))
	ff.DryRun()
	ff.Run(context.Background())
}
