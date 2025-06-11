package main

import (
	"context"
	"fmt"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
)

// ffmpeg -y -i in.mp4 -c:v libx264 -b:v 10M -pass 1 -passlogfile pass_result -f null /dev/null
// ffmpeg -y -i in.mp4 -c:v libx264 -b:v 10M -maxrate 10M -bufsize 10M -pass 2 out2.mp4
func main() {
	var (
		ctx = context.Background()

		ffmpegOpts = []ffmpeg.Option{
			ffmpeg.WithDebug(true), // Enable debug mode
			ffmpeg.WithDry(true),   // Enable dry run mode
		}

		// inputs
		iMain = input.WithSimple("in.mp4")

		// filters
		fDelogo = filter.Scale(-2, -2).Use(iMain.V())
	)

	err := ffmpeg.New(ffmpegOpts...).AddInput(
		iMain,
	).AddFilter(
		fDelogo,
	).AddOutput(
		output.New(
			output.Map(fDelogo),
			output.VideoCodec(codec.X264),
			output.VideoBitrate(10*1024*1024), // 10M
			output.Pass(1),
			output.PassLogfile("pass_result"),
			output.Format("null"),
			output.File("/dev/null"),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}

	err = ffmpeg.New(ffmpegOpts...).AddInput(
		iMain,
	).AddFilter(
		fDelogo,
	).AddOutput(
		output.New(
			output.Map(fDelogo),
			output.VideoCodec(codec.X264),
			output.VideoBitrate(10*1024*1024), // 10M
			output.Maxrate(10*1024*1024),      // 10M
			output.Bufsize(10*1024*1024),      // 10M
			output.Pass(2),
			output.PassLogfile("pass_result"),
			output.File("out.mp4"),
		),
	).Run(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
