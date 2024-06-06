package fusion

import "github.com/google/uuid"

// 轨道元素类型
type TrackItemType string

const (
	// 普通文字
	TrackItemTypeTitle TrackItemType = "title"
	// 高级文字，一般用这个
	TrackItemTypeAdvancedTitle TrackItemType = "advanced_title"
	// 序列帧-文字（气泡文字）
	TrackItemTypeSequenceTitle TrackItemType = "sequence_title"
	// AE文件实现的文字（动效文字）
	TrackItemTypePAGTitle TrackItemType = "pag_title"

	// TODO
	TrackItemTypeFrame TrackItemType = "frame"
	// 字幕文字
	TrackItemTypeSubtitle TrackItemType = "subtitle"
	// TODO
	TrackItemTypeImage TrackItemType = "image"
	// 音频
	TrackItemTypeAudio TrackItemType = "audio"
	// 视频
	TrackItemTypeVideo TrackItemType = "video"
	// Deprecated: 转场
	// 现在转场收到了video中，属于video的一个属性，该属性会影响自身及紧挨着的
	TrackItemTypeTransition TrackItemType = "transition"
)

type TrackItemBase struct {
	Id string `json:"id"` // 根据uuid生成，单个合成协议内唯一
	TimeRange
	Type    TrackItemType `json:"type"`
	AssetId string        `json:"asset_id"` // CME素材id，或本地文件路径
}

// 剪辑时间线
type TimeRange struct {
	StartTime int32 `json:"start_time"`
	Duration  int32 `json:"duration"`
}

// 元素片段
// 音频、视频、特效等元素必填
type Section struct {
	From int32 `json:"from"`
	To   int32 `json:"to"`
}

type Position struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

type ItemSize struct {
	Width  int32 `json:"width,omitempty"`
	Height int32 `json:"height,omitempty"`
}

type ItemContents struct {
	Text      string     `json:"text"`
	TextStyle *TextStyle `json:"text_style,omitempty"`
}

// 文字样式
type TextStyle struct {
	// 通用
	FontSize        int32    `json:"font_size"`
	FontColor       string   `json:"font_color"`
	FontColorList   []string `json:"font_color_list,omitempty"`
	Font            string   `json:"font,omitempty"`
	FontAssetId     string   `json:"font_asset_id,omitempty"`
	FontAlpha       int32    `json:"font_alpha,omitempty"`
	FontBold        int32    `json:"font_bold,omitempty"`
	FontItalic      int32    `json:"font_italic,omitempty"`
	FontUnderline   int32    `json:"font_uline,omitempty"`
	FontAlign       string   `json:"font_align,omitempty"`
	BackgroundColor string   `json:"background_color,omitempty"`
	BackgroundAlpha int32    `json:"background_alpha,omitempty"`
	BorderWidth     int32    `json:"border_width,omitempty"`
	BorderColor     string   `json:"border_color,omitempty"`
	Align           string   `json:"align,omitempty"`

	// 内容填充文字可用
	ShadowColor string `json:"shadow_color,omitempty"`
	ShadowAngle int32  `json:"shadow_angle,omitempty"`
	ShadowAlpha int32  `json:"shadow_alpha,omitempty"`

	// 字幕可用
	Height       int32  `json:"height"`
	Bold         int32  `json:"bold,omitempty"`
	Italic       int32  `json:"italic,omitempty"`
	BottomColor  string `json:"bottom_color,omitempty"`
	BottomAlpha  int32  `json:"bottom_alpha,omitempty"`
	MarginBottom int32  `json:"margin_bottom,omitempty"`

	// 未知
	// LetterSpacing int32      `json:"letter_spacing"`
	// Leading       int32      `json:"leading"`
	// TextBox       *TextBox `json:"text_box"`
}

type TextBox struct {
	Width    float32 `json:"width"`
	Height   float32 `json:"height"`
	FontSize int     `json:"font_size"`
}

func DefaultTextStyle() *TextStyle {
	return &TextStyle{
		FontSize:  60,
		FontColor: "#FFFFFF",
		Align:     "center",
		Height:    220,
		Bold:      0,
		Italic:    0,
	}
}

func DefaultSubtitleStyle() *Style {
	return &Style{
		Id: uuid.NewString(),
		TextStyle: &TextStyle{
			FontSize:  30,
			FontColor: "#FFFFFF",
			Align:     "center",
			Height:    55,
			Bold:      0,
			Italic:    0,
		},
	}
}

func SubtitleStyleWithTextStyle(ts *TextStyle) *Style {
	return &Style{
		Id:        uuid.NewString(),
		TextStyle: ts,
	}
}

// 轨道元素
type TrackItem struct {
	TrackItemBase
	Section *Section `json:"section,omitempty"`
	ItemSize
	Position    *Position     `json:"position,omitempty"`
	TextStyleId string        `json:"text_style_id,omitempty"`
	Contents    *ItemContents `json:"contents,omitempty"`
	Operations  []*Operation  `json:"operations,omitempty"`
	Transition  *Transition   `json:"transition,omitempty"`
}

// 创建轨道
func NewTrackItem(trackItemType TrackItemType) *TrackItem {
	return &TrackItem{
		TrackItemBase: TrackItemBase{
			Id:   uuid.NewString(),
			Type: trackItemType,
		},
	}
}

func (i *TrackItem) SetAssetId(assetId string) *TrackItem {
	i.AssetId = assetId
	return i
}

// 时间线上的起点和持续时间
func (i *TrackItem) SetTimeRange(startTime, duration int32) *TrackItem {
	i.TimeRange = TimeRange{StartTime: startTime, Duration: duration}
	return i
}

// 截取原视频的视频段
func (i *TrackItem) SetSection(from, to int32) *TrackItem {
	i.Section = &Section{From: from, To: to}
	return i
}

func (i *TrackItem) SetItemSize(width, height int32) *TrackItem {
	i.ItemSize = ItemSize{Width: width, Height: height}
	return i
}

func (i *TrackItem) SetPosition(x, y int32) *TrackItem {
	if i.Position != nil {
		i.Position.X = x
		i.Position.Y = y
	} else {
		i.Position = &Position{X: x, Y: y}
	}
	return i
}

func (i *TrackItem) SetSubtitle(text string) *TrackItem {
	// i.TextStyleId = styleId
	i.Contents = &ItemContents{Text: text}
	return i
}

func (i *TrackItem) SetContents(text string, style *TextStyle) *TrackItem {
	i.Contents = &ItemContents{Text: text, TextStyle: style}
	return i
}

func (i *TrackItem) SetOperations(ops ...*Operation) *TrackItem {
	i.Operations = append(i.Operations, ops...)
	return i
}

func (i *TrackItem) SetTransition(ts *Transition) *TrackItem {
	i.Transition = ts
	return i
}
