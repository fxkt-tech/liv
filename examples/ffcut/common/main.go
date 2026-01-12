package main

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffcut/fusion"
	"github.com/fxkt-tech/liv/ffmpeg"
)

func main() {
	proto, err := fusion.New(
		fusion.WithStageSize(960, 540),
		fusion.WithFFmpegOptions(
			ffmpeg.WithDebug(true),
		),
	).
		AppendTrack(
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId("qwer.mp4").
						SetTimeRange(0, 5000).
						SetPosition(100, 200).
						SetItemSize(1280, 720).
						SetSection(0, 5000),
				),
		).
		AppendTrack(
			fusion.NewTrack(fusion.TrackTypeAudio).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeAudio).
						SetAssetId("qwer.wav").
						SetTimeRange(0, 5000).
						SetPosition(100, 200).
						SetItemSize(1280, 720).
						SetSection(0, 5000),
				),
		).
		ExportProto()
	// Export(fusion.ExportConfig{
	// 	Type:    fusion.ExportAudio,
	// 	Outfile: "outout.wav",
	// }) // 导出
	fmt.Println(proto)
	if err != nil {
		fmt.Println(err)
	}
}
