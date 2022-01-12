package main

import (
	"context"
	"fmt"
	"testing"

	"fxkt.tech/ffmpeg"
	"fxkt.tech/ffmpeg/codec"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

func TestNormal(t *testing.T) {
	ff := ffmpeg.Default()
	ff.AddInput(input.New(
		input.SetI("in.mp4"),
	))
	ff.AddOutput(output.New(
		output.SetVideoCoder(codec.VideoX264),
		output.SetAudioCoder(codec.Copy),
		output.SetFile("out.mp4"),
	))
	err := ff.Run(context.Background())
	if err != nil {
		fmt.Println(err)
	}
}
