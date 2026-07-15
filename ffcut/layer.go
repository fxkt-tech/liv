package ffcut

type LengthUnit string

const (
	LengthUnitPixel   LengthUnit = "px"
	LengthUnitPercent LengthUnit = "percent"
)

type Length struct {
	Value float64    `json:"value"`
	Unit  LengthUnit `json:"unit"`
}

type Anchor string

const (
	AnchorTopLeft     Anchor = "top_left"
	AnchorTopRight    Anchor = "top_right"
	AnchorBottomLeft  Anchor = "bottom_left"
	AnchorBottomRight Anchor = "bottom_right"
	AnchorCenter      Anchor = "center"
)

type Geometry struct {
	Anchor Anchor `json:"anchor"`
	X      Length `json:"x"`
	Y      Length `json:"y"`
	Width  Length `json:"width"`
	Height Length `json:"height"`
}

type LayerKind string

const (
	LayerKindImage    LayerKind = "image"
	LayerKindSubtitle LayerKind = "subtitle"
)

type Layer struct {
	ID       ID             `json:"id"`
	Kind     LayerKind      `json:"kind"`
	Range    TimeRange      `json:"range"`
	Image    *ImageLayer    `json:"image,omitempty"`
	Subtitle *SubtitleLayer `json:"subtitle,omitempty"`
}

type ImageLayer struct {
	Source          LocalSource `json:"source"`
	Geometry        Geometry    `json:"geometry"`
	Opacity         float64     `json:"opacity"`
	RotationDegrees float64     `json:"rotation_degrees"`
}

type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

type SubtitleStyle struct {
	FontFamily      string       `json:"font_family,omitempty"`
	Font            *LocalSource `json:"font,omitempty"`
	FontSize        Length       `json:"font_size"`
	Color           string       `json:"color"`
	BackgroundColor string       `json:"background_color,omitempty"`
	Align           TextAlign    `json:"align"`
}

type SubtitleCue struct {
	ID    ID        `json:"id"`
	Range TimeRange `json:"range"`
	Text  string    `json:"text"`
}

type SubtitleLayer struct {
	Region Geometry      `json:"region"`
	Style  SubtitleStyle `json:"style"`
	Cues   []SubtitleCue `json:"cues"`
}
