package main

import (
	"context"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/naming"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

func main() {
	var (
		ctx = context.Background()

		nm     = naming.New()
		input1 = input.WithSimple("in.mp4")
		// input2 = input.WithTime(3, 5, "in2.mp4")

		split1   = filter.Split(nm.Gen(), 2).Use(stream.V(0))
		overlay1 = filter.Logo(nm.Gen(), 50, 100, filter.LogoTopLeft).Use(split1.Get(0), split1.Get(1))
	)

	ffmpeg.New(
		// ffmpeg.Binary("/usr/local/bin/ffmpeg"),
		// ffmpeg.V(ffmpeg.LogLevelError),
		// ffmpeg.V(""),
		ffmpeg.Debug(true),
		// ffmpeg.Dry(true),
	).AddInput(
		input1,
		//  input2,
	).AddFilter(
		split1, overlay1,
	).AddOutput(
		output.New(
			output.Map(overlay1),
			output.Map(stream.Select(0, stream.MayAudio)),
			output.Metadata("comment", "xx"),
			output.VideoCodec(codec.X264),
			output.AudioCodec(codec.Copy),
			output.File("out.mp4"),
		),
	).Run(ctx)
}
