package shelf

import (
	"context"

	"github.com/fxkt-tech/liv/ffmpeg"
	"github.com/fxkt-tech/liv/ffmpeg/codec"
	"github.com/fxkt-tech/liv/ffmpeg/filter"
	"github.com/fxkt-tech/liv/ffmpeg/filter/fsugar"
	"github.com/fxkt-tech/liv/ffmpeg/input"
	"github.com/fxkt-tech/liv/ffmpeg/output"
	"github.com/fxkt-tech/liv/internal/conv"
	"github.com/fxkt-tech/liv/internal/encoding/json"
	"github.com/fxkt-tech/liv/internal/sugar"
	"github.com/google/uuid"
)

// 轨道排序优先级，数字小的靠前
var TrackTypeSortMap = map[TrackType]int{
	TrackTypeTitle:    1,
	TrackTypeFrame:    2,
	TrackTypeSubtitle: 3,
	TrackTypeImage:    4,
	TrackTypeAudio:    5,
	TrackTypeVideo:    6,
}

// 轨道类型
type TrackType string

const (
	// 文字轨道
	TrackTypeTitle TrackType = "title"
	// TODO
	TrackTypeFrame TrackType = "frame"
	// 字幕轨道
	TrackTypeSubtitle TrackType = "subtitle"
	// TODO
	TrackTypeImage TrackType = "image"
	// 音频轨道
	TrackTypeAudio TrackType = "audio"
	// 视频轨道
	TrackTypeVideo TrackType = "video"
)

var (
	TrackTitleAllowdItems    = []TrackItemType{TrackItemTypeTitle, TrackItemTypeAdvancedTitle, TrackItemTypeSequenceTitle, TrackItemTypePAGTitle}
	TrackSubtitleAllowdItems = []TrackItemType{TrackItemTypeSubtitle}
	TrackAudioAllowdItems    = []TrackItemType{TrackItemTypeAudio}
	TrackVideoAllowdItems    = []TrackItemType{TrackItemTypeVideo, TrackItemTypeImage, TrackItemTypeTransition}
)

// 合成协议
type TrackData struct {
	stageWidth  int32
	stageHeight int32

	tracks []*Track

	err error
	ctx context.Context
}

var New = NewTrackData

func NewTrackData(opts ...ShelfOption) *TrackData {
	d := &TrackData{ctx: context.Background()}
	sugar.Range(opts, func(opt ShelfOption) { opt(d) })
	return d
}

// 获取指定轨道，返回nil则表示不存在
func (d *TrackData) GetTrack(trackType TrackType, idx int) *Track {
	i := 0
	for _, track := range d.tracks {
		if track.Type == trackType {
			if i == idx {
				return track
			}
			i++
		}
	}
	return nil
}

// 添加轨道
func (d *TrackData) AddTrack(track *Track) *TrackData {
	if d.err != nil {
		return d
	}
	d.tracks = append(d.tracks, track)
	return d
}

func (d *TrackData) AppendTrack(tracks ...*Track) *TrackData {
	for _, track := range tracks {
		d.AddTrack(track)
	}
	return d
}

// 轨道排序，使用稳定性排序
func (d *TrackData) Sort() {
	n := len(d.tracks)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if TrackTypeSortMap[d.tracks[j].Type] > TrackTypeSortMap[d.tracks[j+1].Type] {
				d.tracks[j], d.tracks[j+1] = d.tracks[j+1], d.tracks[j]
			}
		}
	}
}

func (d *TrackData) MaxDuration() float32 {
	var maxDuration float32
	for _, track := range d.tracks {
		for _, item := range track.Items {
			if sectionTo := conv.MillToF32(item.Selection.StartTime + item.Selection.Duration); maxDuration < sectionTo {
				maxDuration = sectionTo
			}
		}
	}
	return maxDuration
}

func (d *TrackData) Exec(outfile string) error {
	var (
		ff = ffmpeg.New(
			// ffmpeg.WithDry(true),
			ffmpeg.WithDebug(true),
		)

		maxDuration = d.MaxDuration()

		// 辅助性质的背景板，用于视频流
		bg = filter.Color("black", d.stageWidth, d.stageHeight, maxDuration)
		// 辅助性质的背景板，用于音频流
		abg = filter.ANullSrc(maxDuration)

		// 舞台
		stage = bg
		// 音响
		sound = abg
	)
	ff.AddFilter(bg, abg)

	for i := len(d.tracks) - 1; i >= 0; i-- {
		switch d.tracks[i].Type {
		case TrackTypeVideo:
			for _, item := range d.tracks[i].Items {
				var (
					startTime = conv.MillToF32(item.StartTime)
					duration  = conv.MillToF32(item.Duration)
					from      = conv.MillToF32(item.Section.From)
					to        = conv.MillToF32(item.Section.To)
					w, h      = item.Width, item.Height
					x, y      = item.Position.X, item.Position.Y
					speed     = (to - from) / duration
				)

				// 截取当前素材段
				iAsset := input.WithTimeTo(from, to, item.AssetId)
				ff.AddInput(iAsset)

				// 处理视频流
				fScale := filter.Scale(w, h).Use(iAsset.V())
				fSpeed := filter.SetPTS(fsugar.PTSSpeed(speed)).Use(fScale)
				fSetPTS := filter.SetPTS(fsugar.PTSOffset(startTime)).Use(fSpeed)
				fOverlay := filter.OverlayWithEnable(x, y, fsugar.TimeBetween(startTime, startTime+duration)).Use(stage, fSetPTS)

				ff.AddFilter(fScale, fSpeed, fSetPTS, fOverlay)

				stage = fOverlay // 每完成一步的结果就是当前舞台的模样

				// 处理音频流
				fAtempo := filter.ATempo(speed).Use(iAsset.A())
				fADelay := filter.ADelay(startTime).Use(fAtempo)
				fAMix := filter.AMix(2).Use(sound, fADelay)

				ff.AddFilter(fAtempo, fADelay, fAMix)

				sound = fAMix // 每完成一步的结果就是当前音响的效果

			}
		case TrackTypeAudio:
			for _, item := range d.tracks[i].Items {
				var (
					startTime = conv.MillToF32(item.StartTime)
					duration  = conv.MillToF32(item.Duration)
					from      = conv.MillToF32(item.Section.From)
					to        = conv.MillToF32(item.Section.To)
					speed     = (to - from) / duration
				)

				// 截取当前素材段
				iAsset := input.WithTime(startTime, duration, item.AssetId)

				// 处理音频流
				fADelay := filter.ADelay(startTime).Use(iAsset.A())
				fAtempo := filter.ATempo(speed).Use(fADelay)
				fAMix := filter.AMix(2).Use(sound, fADelay)

				ff.AddInput(iAsset)
				ff.AddFilter(fADelay, fAtempo, fAMix)

				// 每完成一步的结果就是当前音响的效果
				sound = fAMix
			}
		case TrackTypeTitle:
		case TrackTypeSubtitle:
		}
	}

	ff.AddOutput(output.New(
		output.Map(stage),
		output.Map(sound),
		output.VideoCodec(codec.X264),
		output.AudioCodec(codec.AAC),
		// output.AudioCodec(codec.Nope),
		output.File(outfile),
	))
	err := ff.Run(d.ctx)
	if err != nil {
		return err
	}

	return nil
}

// 导出合成协议
func (d *TrackData) Export() (string, error) {
	if d.err != nil {
		return "", d.err
	}

	// 处理字幕样式
	for _, track := range d.tracks {
		if track.Type == TrackTypeSubtitle {
			if len(track.Styles) > 0 {
				for _, tItem := range track.Items {
					tItem.TextStyleId = track.Styles[0].Id
				}
			}
		}
	}
	d.Sort()
	return json.ToString(d.tracks), nil
}

// --- 轨道 ---

type TrackBase struct {
	Id   string    `json:"id"` // 根据uuid生成，只要单个合成协议不重复就行
	Type TrackType `json:"type"`
}

type Style struct {
	Id        string     `json:"id"`
	TextStyle *TextStyle `json:"text_style"`
}

// 轨道
// 轨道类型: title、frame、subtitle、image、audio、video
type Track struct {
	TrackBase
	Items  []*TrackItem `json:"items,omitempty"`
	Styles []*Style     `json:"styles,omitempty"`

	allowedTrackItems []TrackItemType `json:"-"`
	err               error           `json:"-"`
}

// 创建轨道
// 暂时仅支持title/subtitle/audio/video
func NewTrack(trackType TrackType) *Track {
	var (
		base = TrackBase{
			Id:   uuid.NewString(),
			Type: trackType,
		}
		allowedTrackItems []TrackItemType
		err               error
	)

	switch trackType {
	case TrackTypeTitle:
		allowedTrackItems = TrackTitleAllowdItems
	case TrackTypeSubtitle:
		allowedTrackItems = TrackSubtitleAllowdItems
	case TrackTypeAudio:
		allowedTrackItems = TrackAudioAllowdItems
	case TrackTypeVideo:
		allowedTrackItems = TrackVideoAllowdItems
	default:
		err = ErrTrackTypeNotFound
	}

	return &Track{
		TrackBase:         base,
		allowedTrackItems: allowedTrackItems,
		err:               err,
	}
}

// 向轨道中添加元素
func (t *Track) Push(trackItem *TrackItem) *Track {
	if t.err != nil {
		return t
	}

	if !sugar.In(t.allowedTrackItems, trackItem.Type) {
		t.err = ErrTrackItemTypeNotMatch
		return t
	}

	t.Items = append(t.Items, trackItem)
	return t
}

func (t *Track) Append(trackItems ...*TrackItem) *Track {
	for _, trackItem := range trackItems {
		t.Push(trackItem)
	}
	return t
}

func (t *Track) SetStyles(styles ...*Style) *Track {
	t.Styles = append(t.Styles, styles...)
	return t
}
