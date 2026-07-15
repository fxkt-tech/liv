package ffvmix

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffcut"
)

type LayerKind = ffcut.LayerKind

const (
	LayerKindImage    = ffcut.LayerKindImage
	LayerKindSubtitle = ffcut.LayerKindSubtitle
)

type LayerSpec struct {
	ID       ID                 `json:"id"`
	Kind     LayerKind          `json:"kind"`
	Timing   LayerTiming        `json:"timing"`
	Image    *ImageLayerSpec    `json:"image,omitempty"`
	Subtitle *SubtitleLayerSpec `json:"subtitle,omitempty"`
}

type LayerTimeKind string

const (
	LayerTimeFullOutput LayerTimeKind = "full_output"
	LayerTimeSlot       LayerTimeKind = "slot"
	LayerTimeAbsolute   LayerTimeKind = "absolute"
)

type LayerTiming struct {
	Kind       LayerTimeKind        `json:"kind"`
	FullOutput *FullOutputTiming    `json:"full_output,omitempty"`
	Slot       *SlotTiming          `json:"slot,omitempty"`
	Absolute   *AbsoluteLayerTiming `json:"absolute,omitempty"`
}

type FullOutputTiming struct{}

type SlotTiming struct {
	SlotID   ID              `json:"slot_id"`
	Offset   ffcut.Duration  `json:"offset,omitempty"`
	Duration *ffcut.Duration `json:"duration,omitempty"`
}

type AbsoluteLayerTiming struct {
	Range ffcut.TimeRange `json:"range"`
}

type ImageLayerSpec struct {
	Path            string         `json:"path"`
	Geometry        ffcut.Geometry `json:"geometry"`
	Opacity         float64        `json:"opacity"`
	RotationDegrees float64        `json:"rotation_degrees"`
}

type ImageLayerConfig struct {
	Timing          LayerTiming
	Path            string
	Geometry        ffcut.Geometry
	Opacity         *float64
	RotationDegrees float64
}

type SubtitleInputKind string

const (
	SubtitleInputStructured SubtitleInputKind = "structured"
	SubtitleInputSRT        SubtitleInputKind = "srt"
	SubtitleInputASS        SubtitleInputKind = "ass"
)

type SubtitleInput struct {
	Kind       SubtitleInputKind       `json:"kind"`
	Structured *StructuredSubtitleSpec `json:"structured,omitempty"`
	SRT        *SubtitleFileSpec       `json:"srt,omitempty"`
	ASS        *SubtitleFileSpec       `json:"ass,omitempty"`
}

type StructuredSubtitleSpec struct {
	Cues []SubtitleCueSpec `json:"cues"`
}

type SubtitleCueSpec struct {
	ID    ID              `json:"id"`
	Range ffcut.TimeRange `json:"range"`
	Text  string          `json:"text"`
}

type SubtitleFileSpec struct {
	Path string `json:"path"`
}

type SubtitleStyleSpec struct {
	FontFamily      string          `json:"font_family,omitempty"`
	FontPath        string          `json:"font_path,omitempty"`
	FontSize        ffcut.Length    `json:"font_size"`
	Color           string          `json:"color"`
	BackgroundColor string          `json:"background_color,omitempty"`
	Align           ffcut.TextAlign `json:"align"`
}

type SubtitleLayerSpec struct {
	Region ffcut.Geometry    `json:"region"`
	Style  SubtitleStyleSpec `json:"style"`
	Input  SubtitleInput     `json:"input"`
}

type SubtitleLayerConfig struct {
	Timing LayerTiming
	Region ffcut.Geometry
	Style  SubtitleStyleSpec
	Input  SubtitleInput
}

func FullOutputLayerTiming() LayerTiming {
	return LayerTiming{Kind: LayerTimeFullOutput, FullOutput: &FullOutputTiming{}}
}

func SlotLayerTiming(slotID ID, offset ffcut.Duration, duration *ffcut.Duration) LayerTiming {
	return LayerTiming{
		Kind: LayerTimeSlot,
		Slot: &SlotTiming{SlotID: slotID, Offset: offset, Duration: cloneDuration(duration)},
	}
}

func AbsoluteLayerRange(value ffcut.TimeRange) LayerTiming {
	return LayerTiming{
		Kind:     LayerTimeAbsolute,
		Absolute: &AbsoluteLayerTiming{Range: value},
	}
}

func StructuredSubtitles(cues []SubtitleCueSpec) SubtitleInput {
	cloned := append([]SubtitleCueSpec(nil), cues...)
	for index := range cloned {
		if cloned[index].ID == "" {
			cloned[index].ID = newID("cue")
		}
	}
	return SubtitleInput{
		Kind:       SubtitleInputStructured,
		Structured: &StructuredSubtitleSpec{Cues: cloned},
	}
}

func SRTSubtitles(path string) SubtitleInput {
	return SubtitleInput{Kind: SubtitleInputSRT, SRT: &SubtitleFileSpec{Path: path}}
}

func ASSSubtitles(path string) SubtitleInput {
	return SubtitleInput{Kind: SubtitleInputASS, ASS: &SubtitleFileSpec{Path: path}}
}

func (t *Template) AddImageLayer(config ImageLayerConfig) (*LayerSpec, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: template is nil", ErrInvalidTemplate)
	}
	opacity := 1.0
	if config.Opacity != nil {
		opacity = *config.Opacity
	}
	layer := &LayerSpec{
		ID:     newID("layer"),
		Kind:   LayerKindImage,
		Timing: cloneLayerTiming(config.Timing),
		Image: &ImageLayerSpec{
			Path:            config.Path,
			Geometry:        config.Geometry,
			Opacity:         opacity,
			RotationDegrees: config.RotationDegrees,
		},
	}
	t.Layers = append(t.Layers, layer)
	return layer, nil
}

func (t *Template) AddSubtitleLayer(config SubtitleLayerConfig) (*LayerSpec, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: template is nil", ErrInvalidTemplate)
	}
	layer := &LayerSpec{
		ID:     newID("layer"),
		Kind:   LayerKindSubtitle,
		Timing: cloneLayerTiming(config.Timing),
		Subtitle: &SubtitleLayerSpec{
			Region: config.Region,
			Style:  config.Style,
			Input:  cloneSubtitleInput(config.Input),
		},
	}
	t.Layers = append(t.Layers, layer)
	return layer, nil
}

func cloneLayerTiming(value LayerTiming) LayerTiming {
	cloned := LayerTiming{Kind: value.Kind}
	if value.FullOutput != nil {
		cloned.FullOutput = &FullOutputTiming{}
	}
	if value.Slot != nil {
		cloned.Slot = &SlotTiming{
			SlotID:   value.Slot.SlotID,
			Offset:   value.Slot.Offset,
			Duration: cloneDuration(value.Slot.Duration),
		}
	}
	if value.Absolute != nil {
		cloned.Absolute = &AbsoluteLayerTiming{Range: value.Absolute.Range}
	}
	return cloned
}

func cloneSubtitleInput(value SubtitleInput) SubtitleInput {
	cloned := SubtitleInput{Kind: value.Kind}
	if value.Structured != nil {
		cloned.Structured = &StructuredSubtitleSpec{
			Cues: append([]SubtitleCueSpec(nil), value.Structured.Cues...),
		}
	}
	if value.SRT != nil {
		cloned.SRT = &SubtitleFileSpec{Path: value.SRT.Path}
	}
	if value.ASS != nil {
		cloned.ASS = &SubtitleFileSpec{Path: value.ASS.Path}
	}
	return cloned
}
