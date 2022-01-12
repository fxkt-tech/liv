package main

import (
	"testing"

	"fxkt.tech/ffmpeg"
	"fxkt.tech/ffmpeg/codec"
	"fxkt.tech/ffmpeg/filter"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

func TestFilter(t *testing.T) {
	ff := ffmpeg.Default()
	ff.AddInput(input.New(
		input.SetI("in.mp4"),
	))
	ff.AddFilter(filter.New(
		filter.SetContent(filter.Logo(10, 10, filter.LogoTopRight)),
	))
	ff.AddOutput(output.New(
		output.SetVideoCoder(codec.VideoX264),
		output.SetAudioCoder(codec.Copy),
		output.SetFile("out.mp4"),
	))
	ff.DryRun()
}
