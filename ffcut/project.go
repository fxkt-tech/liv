// Package ffcut defines the versioned, renderer-independent FFcut timeline
// protocol.
package ffcut

const ProjectVersion = 2

type ID string

type Project struct {
	Version  int          `json:"version"`
	ID       ID           `json:"id"`
	Canvas   Canvas       `json:"canvas"`
	Video    Sequence     `json:"video"`
	Audio    []AudioTrack `json:"audio,omitempty"`
	Layers   []Layer      `json:"layers,omitempty"`
	Metadata Metadata     `json:"metadata"`
}

type FrameRate struct {
	Numerator   int32 `json:"numerator"`
	Denominator int32 `json:"denominator"`
}

type Canvas struct {
	Width      int32      `json:"width"`
	Height     int32      `json:"height"`
	FrameRate  FrameRate  `json:"frame_rate"`
	Background Background `json:"background"`
}

type BackgroundKind string

const (
	BackgroundKindColor BackgroundKind = "color"
	BackgroundKindImage BackgroundKind = "image"
	BackgroundKindBlur  BackgroundKind = "blur"
)

type Background struct {
	Kind  BackgroundKind   `json:"kind"`
	Color *ColorBackground `json:"color,omitempty"`
	Image *ImageBackground `json:"image,omitempty"`
	Blur  *BlurBackground  `json:"blur,omitempty"`
}

type ColorBackground struct {
	Color string `json:"color"`
}

type ImageBackground struct {
	Source LocalSource `json:"source"`
	Fit    FitMode     `json:"fit"`
}

type BlurBackground struct {
	Sigma float64 `json:"sigma"`
}

type FitMode string

const (
	FitModeCover   FitMode = "cover"
	FitModeContain FitMode = "contain"
	FitModeStretch FitMode = "stretch"
)

type LocalSource struct {
	ID ID `json:"id"`
	// Path is an absolute path resolved before Project construction.
	Path        string           `json:"path"`
	Fingerprint MediaFingerprint `json:"fingerprint"`
}

type MediaFingerprint struct {
	Size             int64  `json:"size"`
	ModifiedUnixNano int64  `json:"modified_unix_nano"`
	SHA256           string `json:"sha256,omitempty"`
}

type Metadata struct {
	TemplateFingerprint    string             `json:"template_fingerprint"`
	Seed                   uint64             `json:"seed"`
	CombinationFingerprint string             `json:"combination_fingerprint"`
	Selections             []Selection        `json:"selections"`
	Constraints            []ConstraintRecord `json:"constraints,omitempty"`
}

type SelectionKind string

const (
	SelectionKindVideo      SelectionKind = "video"
	SelectionKindTransition SelectionKind = "transition"
	SelectionKindBGM        SelectionKind = "bgm"
)

type Selection struct {
	Kind             SelectionKind `json:"kind"`
	DimensionID      ID            `json:"dimension_id"`
	OptionID         ID            `json:"option_id"`
	AssetFingerprint string        `json:"asset_fingerprint,omitempty"`
}

type ConstraintRecord struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
}
