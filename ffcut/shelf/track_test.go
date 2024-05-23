package shelf_test

import (
	"fmt"
	"testing"

	"github.com/fxkt-tech/liv/ffcut/shelf"
)

func TestExport(t *testing.T) {
	trackData, err := shelf.New(
		shelf.WithStageSize(540, 960),
	).
		AddTrack(
			// 添加视频轨
			shelf.NewTrack(shelf.TrackTypeVideo).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId("660521dc64b43b0001ce490a").
						SetSelection(0, 5000).
						SetSection(0, 5000).
						SetItemSize(540, 960).
						SetPosition(270, 480),
				)).
		AddTrack(
			// 添加文字轨
			shelf.NewTrack(shelf.TrackTypeTitle).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeSequenceTitle).
						SetAssetId("608bc4d689ea7200013ff242@Public@CME").
						SetSelection(0, 5000).
						SetPosition(270, 180).
						SetContents("超好吃！", &shelf.TextStyle{
							Font:      "huakangshaonvwenziW5-2",
							FontSize:  40,
							FontColor: "#173563",
							Align:     "center",
						}),
				),
		).
		AddTrack(
			// 添加音频轨
			shelf.NewTrack(shelf.TrackTypeAudio).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeAudio).
						SetAssetId("65df048fb95d8900015302c9").
						SetSelection(0, 5000).
						SetSection(0, 5000),
				),
		).
		AddTrack(
			// 添加字幕轨
			shelf.NewTrack(shelf.TrackTypeSubtitle).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeSubtitle).
						SetSelection(0, 1000).
						SetSubtitle("此处字幕1"),
					shelf.NewTrackItem(shelf.TrackItemTypeSubtitle).
						SetSelection(1000, 3000).
						SetSubtitle("此处字幕2"),
					shelf.NewTrackItem(shelf.TrackItemTypeSubtitle).
						SetSelection(4000, 1000).
						SetSubtitle("此处字幕3"),
				).
				SetStyles(shelf.DefaultSubtitleStyle()),
		).
		Export() // 导出
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}

func TestExec(t *testing.T) {
	err := shelf.New(
		// shelf.WithStageSize(540, 960),
		shelf.WithStageSize(1920, 1080),
	).
		AppendTrack(
			// 添加视频轨
			shelf.NewTrack(shelf.TrackTypeVideo).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(0, 10000).
						SetSection(0, 10000).
						SetItemSize(960, 540).
						SetPosition(600, 600),
				),
			shelf.NewTrack(shelf.TrackTypeVideo).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(8000, 2000).
						SetSection(1000, 3000).
						SetItemSize(1920, 1080).
						SetPosition(0, 0),
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(20000, 2000).
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
	err := shelf.New(
		// shelf.WithStageSize(540, 960),
		shelf.WithStageSize(1080, 1920),
	).
		AppendTrack(
			// 添加视频轨
			// shelf.NewTrack(shelf.TrackTypeVideo).
			// 	Append(
			// 		shelf.NewTrackItem(shelf.TrackItemTypeVideo).
			// 			SetAssetId(`in.mp4`).
			// 			SetSelection(0, 10000).
			// 			SetSection(0, 10000).
			// 			SetItemSize(960, 540).
			// 			SetPosition(600, 600),
			// 	),
			shelf.NewTrack(shelf.TrackTypeVideo).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(filename).
						SetSelection(0, 4760).
						SetSection(0, 5950).
						SetItemSize(1080, 1920).
						SetPosition(0, 0),
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(filename).
						SetSelection(4760, 7960).
						SetSection(5920, 13880).
						SetItemSize(1080, 1920).
						SetPosition(0, 0),
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(filename).
						SetSelection(12720, 6120).
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

func TestSpeed(t *testing.T) {
	err := shelf.New(
		shelf.WithStageSize(540, 960),
		// shelf.WithStageSize(1080, 1920),
	).
		AppendTrack(
			// 添加视频轨
			shelf.NewTrack(shelf.TrackTypeVideo).
				Append(
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(0, 2727).
						SetSection(0, 3000).
						SetItemSize(540, 960).
						SetPosition(0, 0),
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(2000, 4000).
						SetSection(3000, 7000).
						SetItemSize(540, 960).
						SetPosition(0, 0),
					shelf.NewTrackItem(shelf.TrackItemTypeVideo).
						SetAssetId(`in.mp4`).
						SetSelection(6000, 2727).
						SetSection(7000, 10000).
						SetItemSize(540, 960).
						SetPosition(0, 0),
					// shelf.NewTrackItem(shelf.TrackItemTypeVideo).
					// 	SetAssetId(`in.mp4`).
					// 	SetSelection(0, 3000).
					// 	SetSection(0, 3000).
					// 	SetItemSize(540, 960).
					// 	SetPosition(0, 0),
					// shelf.NewTrackItem(shelf.TrackItemTypeVideo).
					// 	SetAssetId(`in.mp4`).
					// 	SetSelection(3000, 4000).
					// 	SetSection(3000, 7000).
					// 	SetItemSize(540, 960).
					// 	SetPosition(0, 0),
					// shelf.NewTrackItem(shelf.TrackItemTypeVideo).
					// 	SetAssetId(`in.mp4`).
					// 	SetSelection(7000, 3000).
					// 	SetSection(7000, 10000).
					// 	SetItemSize(540, 960).
					// 	SetPosition(0, 0),
				),
		).
		Exec("out_test_1_1.mp4") // 导出
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(trackData)
	// t.Log(json.Pretty([]byte(trackData)))
}
