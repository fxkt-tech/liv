package fusion_test

import (
	"fmt"
	"testing"

	"github.com/fxkt-tech/liv/ffcut/fusion"
)

func TestExport(t *testing.T) {
	trackData, err := fusion.New(
		fusion.WithStageSize(540, 960),
	).
		AddTrack(
			// 添加视频轨
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId("660521dc64b43b0001ce490a").
						SetTimeRange(0, 5000).
						SetSection(0, 5000).
						SetItemSize(540, 960).
						SetPosition(270, 480),
				)).
		AddTrack(
			// 添加文字轨
			fusion.NewTrack(fusion.TrackTypeTitle).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeSequenceTitle).
						SetAssetId("608bc4d689ea7200013ff242@Public@CME").
						SetTimeRange(0, 5000).
						SetPosition(270, 180).
						SetContents("超好吃！", &fusion.TextStyle{
							Font:      "huakangshaonvwenziW5-2",
							FontSize:  40,
							FontColor: "#173563",
							Align:     "center",
						}),
				),
		).
		AddTrack(
			// 添加音频轨
			fusion.NewTrack(fusion.TrackTypeAudio).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeAudio).
						SetAssetId("65df048fb95d8900015302c9").
						SetTimeRange(0, 5000).
						SetSection(0, 5000),
				),
		).
		AddTrack(
			// 添加字幕轨
			fusion.NewTrack(fusion.TrackTypeSubtitle).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeSubtitle).
						SetTimeRange(0, 1000).
						SetSubtitle("此处字幕1"),
					fusion.NewTrackItem(fusion.TrackItemTypeSubtitle).
						SetTimeRange(1000, 3000).
						SetSubtitle("此处字幕2"),
					fusion.NewTrackItem(fusion.TrackItemTypeSubtitle).
						SetTimeRange(4000, 1000).
						SetSubtitle("此处字幕3"),
				).
				SetStyles(fusion.DefaultSubtitleStyle()),
		).
		Export() // 导出
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}

func TestExec(t *testing.T) {
	err := fusion.New(
		// fusion.WithStageSize(540, 960),
		fusion.WithStageSize(1920, 1080),
	).
		AppendTrack(
			// 添加视频轨
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(0, 10000).
						SetSection(0, 10000).
						SetItemSize(960, 540).
						SetPosition(600, 600),
				),
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(8000, 2000).
						SetSection(1000, 3000).
						SetItemSize(1920, 1080).
						SetPosition(0, 0),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(20000, 2000).
						SetSection(3000, 5000).
						SetItemSize(1920, 1080).
						SetPosition(100, 50),
				),
		).
		Exec("out_test.mp4") // 导出
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}

func TestExec2(t *testing.T) {
	filename := "/Users/justyer/Desktop/in.mp4"
	err := fusion.New(
		// fusion.WithStageSize(540, 960),
		fusion.WithStageSize(1080, 1920),
	).
		AppendTrack(
			// 添加视频轨
			// fusion.NewTrack(fusion.TrackTypeVideo).
			// 	Append(
			// 		fusion.NewTrackItem(fusion.TrackItemTypeVideo).
			// 			SetAssetId(`in.mp4`).
			// 			SetTimeRange(0, 10000).
			// 			SetSection(0, 10000).
			// 			SetItemSize(960, 540).
			// 			SetPosition(600, 600),
			// 	),
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(filename).
						SetTimeRange(0, 4760).
						SetSection(0, 5950).
						SetItemSize(1080, 1920).
						SetPosition(0, 0),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(filename).
						SetTimeRange(4760, 7960).
						SetSection(5920, 13880).
						SetItemSize(1080, 1920).
						SetPosition(0, 0),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(filename).
						SetTimeRange(12720, 6120).
						SetSection(13880, 20000).
						SetItemSize(1080, 1920).
						SetPosition(0, 0),
				),
		).
		Exec("out_test.mp4") // 导出
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}

func F32ToMill(f float32) int32 {
	return int32(f * 1000)
}

func TestSpeed(t *testing.T) {
	var (
		duration = float32(10)

		speed = float32(1.2)

		from1, to1 = float32(0), duration * 0.3
		from2, to2 = to1, duration * 0.7
		from3, to3 = to2, duration

		st1, d1 = float32(0), (to1 - from1) / speed
		st2, d2 = d1, to2 - from2
		st3, d3 = d1 + d2, (to3 - from3) / speed
	)
	err := fusion.New(
		fusion.WithStageSize(540, 960),
	).
		AppendTrack(
			// 添加图片轨
			// fusion.NewTrack(fusion.TrackTypeVideo).
			// 	Append(
			// 		fusion.NewTrackItem(fusion.TrackItemTypeImage).
			// 			SetAssetId(`/Users/justyer/go/src/github.com/fxkt-tech/liv/fftool/mask.png`).
			// 			SetTimeRange(0, 2800*2+4000).
			// 			SetItemSize(540, 960).
			// 			SetPosition(0, 0),
			// 	),
			// 添加视频轨
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(F32ToMill(st1), F32ToMill(d1)).
						SetSection(F32ToMill(from1), F32ToMill(to1)).
						SetItemSize(540, 960).
						SetPosition(0, 0),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(F32ToMill(st2), F32ToMill(d2)).
						SetSection(F32ToMill(from2), F32ToMill(to2)).
						SetItemSize(540, 960).
						SetPosition(0, 0),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetTimeRange(F32ToMill(st3), F32ToMill(d3)).
						SetSection(F32ToMill(from3), F32ToMill(to3)).
						SetItemSize(540, 960).
						SetPosition(0, 0),
				),
		).
		Exec("out_test_1_1.mp4") // 导出
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}

func TestVmix(t *testing.T) {
	err := fusion.New(
		fusion.WithStageSize(853, 480),
	).
		AppendTrack(
			fusion.NewTrack(fusion.TrackTypeVideo).
				Append(
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId("/Users/justyer/Desktop/fade1.mp4").
						SetTimeRange(0, 2500).
						SetSection(0, 2500).
						SetItemSize(853, 480).
						SetPosition(0, 0).
						SetTransition(&fusion.Transition{
							Name:      "fade",
							Duration:  1000,
							Color:     "black",
							WithAudio: true,
						}),
					fusion.NewTrackItem(fusion.TrackItemTypeVideo).
						SetAssetId("/Users/justyer/Desktop/fade2.mp4").
						SetTimeRange(2500, 2500).
						SetSection(0, 2500).
						SetItemSize(853, 480).
						SetPosition(0, 0),
				),
		).
		Exec("out_test.mp4") // 导出
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}
