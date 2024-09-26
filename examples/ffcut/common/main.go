package main

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffcut/fusion"
	"github.com/fxkt-tech/liv/ffmpeg"
)

func main() {
	err := fusion.New(
		fusion.WithStageSize(960, 540),
		fusion.WithFFmpegOptions(
			ffmpeg.WithDebug(true),
		),
	).
		AppendTrack(
			fusion.NewTrack(fusion.TrackTypeAudio).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeAudio).
						SetAssetId("/Users/justyer/Desktop/qwer.wav").
						SetTimeRange(0, 5000).
						SetSection(0, 5000),
				),
		).
		Export(fusion.ExportConfig{
			Type:    fusion.ExportAudio,
			Outfile: "outout.wav",
		}) // 导出
	if err != nil {
		fmt.Println(err)
	}
}
