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
		output.VideoCodec(codec.X264),
		output.AudioCodec(codec.Copy),
		output.Map(filter.SelectStream(0, filter.StreamVideo, true)),
		output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
		output.File("m.m3u8"),
		output.MovFlags("faststart"),
		output.HlsSegmentType("mpegts"),
		output.HlsFlags("independent_segments"),
		output.HlsPlaylistType("vod"),
		output.HlsTime(2),
		// output.HlsKeyInfoFile("file.keyinfo"), // 加密
		output.HlsSegmentFilename("m-%5d.ts"),
		output.Format(codec.Hls),
	))
	ff.DryRun()
	ff.Run(context.Background())
}

func TestAdaptiveHLS(t *testing.T) {
	folder := "video/"
	ff := ffmpeg.Default()
	ff.LogLevel("error")
	ff.AddInput(input.New(
		input.I("in.mp4"),
	))
	ff.AddFilter(filter.New(
		filter.InStream(filter.SelectStream(0, filter.StreamVideo, true)),
		filter.Content(filter.Scale(-2, 360)),
		filter.OutStream("360P"),
	), filter.New(
		filter.InStream(filter.SelectStream(0, filter.StreamVideo, true)),
		filter.Content(filter.Scale(-2, 540)),
		filter.OutStream("540P"),
	))
	ff.AddOutput(output.New(
		output.VideoCodec(codec.X264),
		output.AudioCodec(codec.Copy),
		output.Map("360P"),
		output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
		output.Map("540P"),
		output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
		output.VarStreamMap("v:0,a:0 v:1,a:1"),
		output.MovFlags("faststart"),
		output.HlsSegmentType("mpegts"),
		output.HlsFlags("independent_segments"),
		output.HlsPlaylistType("vod"),
		output.HlsTime(2),
		// output.HlsKeyInfoFile("file.keyinfo"), // 加密
		output.MasterPlName("m.m3u8"),
		output.HlsSegmentFilename(folder+"%v/%5d.ts"),
		output.File(folder+"%v/r.m3u8"),
		output.Format(codec.Hls),
	))
	ff.DryRun()
	ff.Run(context.Background())
}
