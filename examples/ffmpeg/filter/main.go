package main

import (
	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

func main() {
	var (
		nm     = naming.New()
		input1 = input.WithSimple("in.mp4")
		input2 = input.WithTime(3, 5, "in2.mp4")

		scale1   = filter.Scale(nm.Gen(), 0, 0)
		split1   = filter.Split(nm.Gen(), 2).Use(scale1)
		overlay1 = filter.Logo(nm.Gen(), 10, 10, filter.LogoTopRight).Use(scale1, split1)
	)

	ffmpeg.NewFFmpeg(
		ffmpeg.Binary("/usr/local/bin/ffmpeg"),
		ffmpeg.V(ffmpeg.LogLevelError),
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
