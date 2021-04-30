package test

import (
	"testing"

	"fxkt.tech/ffmpeg"
)

func TestExample(t *testing.T) {
	ff := ffmpeg.NewFFmpeg()
	ff.AddInputs(
		"video.mp4",
		"logo.png",
	)
	ff.AddFilter(
		ffmpeg.NewFilter("dl", `delogo=0:0:400:200`, "0"),
		ffmpeg.NewFilter("d_480][d_360", `split=2`, "dl"),
		ffmpeg.NewFilter("rlt_480", `scale=trunc(oh*a/2)*2:480`, "d_480"),
		ffmpeg.NewFilter("rlt_360", `scale=trunc(oh*a/2)*2:360`, "d_360"),
	)
	ff.OutputGraph(
		ffmpeg.NewOutput(
			"-map", "[rlt_480]",
			"-map", "0:a",
			"-metadata", "comment=fu789sg",
			"-c:v", "libx264", "-c:a", "copy",
			"-threads", "4",
			"-max_muxing_queue_size", "4086",
			"-movflags", "faststart",
			"out_480.mp4",
		),
		ffmpeg.NewOutput(
			"-map", "[rlt_360]",
			"-map", "0:a",
			"-metadata", "comment=fu789sg",
			"-c:v", "libx264", "-c:a", "copy",
			"-threads", "4",
			"-max_muxing_queue_size", "4086",
			"-movflags", "faststart",
			"out_480.mp4",
		),
	)
	err := ff.Run()
	if err != nil {
		t.Fatal(err)
	}
}
