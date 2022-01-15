package ffmpeg

import (
	"testing"

	"fxkt.tech/ffmpeg/filter"
	"fxkt.tech/ffmpeg/input"
	"fxkt.tech/ffmpeg/output"
)

func TestFFmpeg(t *testing.T) {
	ff := Default()
	ff.AddInput(
		input.New(
			input.I("vieo.mp4"),
		),
	)
	ff.AddFilter(
		filter.New(
			filter.InStream("0"),
			filter.Content("scale=trunc(oh*a/2)*2:720"),
			filter.OutStream("x720"),
		),
		filter.New(
			filter.InStream("x720"),
			filter.Content("delogo=0:0:100:100"),
			filter.OutStream("xx720"),
		),
	)
	ff.AddOutput(
		output.New(
			output.Map("xx720"),
			output.Map("0:a"),
			output.Metadata("comment", "yilan888"),
			output.File("out720.mp4"),
		),
	)
	ff.DryRun()
}
