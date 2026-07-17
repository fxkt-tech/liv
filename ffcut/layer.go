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
	LayerKindMedia    LayerKind = "media"
	LayerKindSubtitle LayerKind = "subtitle"
)

type Layer struct {
	ID       ID             `json:"id"`
	Kind     LayerKind      `json:"kind"`
	Range    TimeRange      `json:"range"`
	Image    *ImageLayer    `json:"image,omitempty"`
	Media    *MediaLayer    `json:"media,omitempty"`
	Subtitle *SubtitleLayer `json:"subtitle,omitempty"`
}

type ImageLayer struct {
	Source          LocalSource `json:"source"`
	Geometry        Geometry    `json:"geometry"`
	Opacity         float64     `json:"opacity"`
	RotationDegrees float64     `json:"rotation_degrees"`
}

type MediaKind string

const (
	MediaKindImage     MediaKind = "image"
	MediaKindAnimation MediaKind = "animation"
	MediaKindVideo     MediaKind = "video"
)

type MediaLayer struct {
	Kind            MediaKind   `json:"kind"`
	Source          LocalSource `json:"source"`
	Geometry        Geometry    `json:"geometry"`
	Opacity         float64     `json:"opacity"`
	RotationDegrees float64     `json:"rotation_degrees"`
	Loop            bool        `json:"loop"`
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
	StrokeColor     string       `json:"stroke_color,omitempty"`
	StrokeWidth     Length       `json:"stroke_width,omitempty"`
}

type SubtitleCue struct {
	ID    ID        `json:"id"`
	Range TimeRange `json:"range"`
	Text  string    `json:"text"`
}

type SubtitleLayer struct {
	Region          Geometry      `json:"region"`
	Style           SubtitleStyle `json:"style"`
	Cues            []SubtitleCue `json:"cues"`
	Opacity         *float64      `json:"opacity,omitempty"`
	RotationDegrees float64       `json:"rotation_degrees,omitempty"`
}
