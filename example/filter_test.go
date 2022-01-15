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
		input.I("in.mp4"),
	))
	ff.AddFilter(filter.New(
		filter.Content(filter.Logo(10, 10, filter.LogoTopRight)),
	))
	ff.AddOutput(output.New(
		output.VideoCoder(codec.VideoX264),
		output.AudioCoder(codec.Copy),
		output.File("out.mp4"),
	))
	ff.DryRun()
}
