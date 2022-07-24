package main

import (
	echotool "fxkt.tech/echo/tool"
	"fxkt.tech/echo/tool/codec"
	"fxkt.tech/echo/tool/filter"
	"fxkt.tech/echo/tool/input"
	"fxkt.tech/echo/tool/output"
)

func main() {
	var (
		input1 = input.WithSimple("in.mp4")
		input2 = input.WithTime(3, 5, "in2.mp4")

		scale1   = filter.Scale("scale1", 0, 0)
		split1   = filter.Split("split", 2).Use(scale1)
		overlay1 = filter.Logo("overlay1", 10, 10, filter.LogoTopRight).Use(scale1, split1)
	)

	echotool.NewFFmpeg(
		echotool.CmdLoc("/usr/local/bin/ffmpeg"),
		echotool.LogLevel("error"),
	).AddInput(
		input1, input2,
	).AddFilter(
		scale1, split1, overlay1,
	).AddOutput(
		output.New(
			output.Map(overlay1.Name(0)),
			output.Map(filter.SelectStream(0, filter.StreamAudio, false)),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out.mp4"),
		),
	).DryRun()
}
