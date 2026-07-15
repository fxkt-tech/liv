package ffvmix

import (
	"github.com/fxkt-tech/liv/ffcut"
)

type CompiledAsset struct {
	Path        string
	Fingerprint ffcut.MediaFingerprint
}

func (a CompiledAsset) FingerprintString() string {
	return assetFingerprint(a)
}

type VideoMetadata struct {
	Width         int32
	Height        int32
	Duration      ffcut.Duration
	HasAudio      bool
	AudioDuration ffcut.Duration
}

type CompiledVideo struct {
	ID        ID
	Asset     CompiledAsset
	Metadata  VideoMetadata
	Weight    float64
	Fit       ffcut.FitMode
	AudioGain float64
	Plan      AdaptationPlan
}

type CompiledSlot struct {
	ID                ID
	Name              string
	HasTargetDuration bool
	TargetDuration    ffcut.Duration
	Policy            EffectiveSlotPolicy
	Videos            []CompiledVideo
}

type CompiledTransition struct {
	ID             ID
	Kind           ffcut.TransitionKind
	Duration       ffcut.Duration
	AudioCrossfade bool
	Weight         float64
}

type compatibilityKey struct {
	TransitionID ID
	FromVideoID  ID
	ToVideoID    ID
}

type CompiledJoin struct {
	ID            ID
	FromSlotID    ID
	ToSlotID      ID
	Transitions   []CompiledTransition
	compatibility map[compatibilityKey]bool
}

func (j CompiledJoin) IsCompatible(transitionID, fromVideoID, toVideoID ID) bool {
	return j.compatibility[compatibilityKey{
		TransitionID: transitionID,
		FromVideoID:  fromVideoID,
		ToVideoID:    toVideoID,
	}]
}

type CompiledBGM struct {
	ID            ID
	Asset         CompiledAsset
	SourceRange   ffcut.TimeRange
	TimelineStart ffcut.Duration
	Loop          bool
	Gain          float64
	FadeIn        ffcut.Duration
	FadeOut       ffcut.Duration
	Weight        float64
	templateIndex int
}

type CompiledBackground struct {
	Kind  BackgroundKind
	Color *ColorBackgroundSpec
	Image *CompiledImageBackground
	Blur  *BlurBackgroundSpec
}

type CompiledImageBackground struct {
	Asset CompiledAsset
	Fit   ffcut.FitMode
}

type NormalizedCue struct {
	ID    ID
	Range ffcut.TimeRange
	Text  string
}

type CompiledSubtitleStyle struct {
	FontFamily      string
	Font            *CompiledAsset
	FontSize        ffcut.Length
	Color           string
	BackgroundColor string
	Align           ffcut.TextAlign
}

type CompiledImageLayer struct {
	Asset           CompiledAsset
	Geometry        ffcut.Geometry
	Opacity         float64
	RotationDegrees float64
}

type CompiledSubtitleLayer struct {
	Region ffcut.Geometry
	Style  CompiledSubtitleStyle
	Cues   []NormalizedCue
}

type CompiledLayer struct {
	ID       ID
	Kind     LayerKind
	Timing   LayerTiming
	Image    *CompiledImageLayer
	Subtitle *CompiledSubtitleLayer
}

type CompiledTemplate struct {
	id          ID
	fingerprint string
	canvas      CanvasSpec
	background  CompiledBackground
	defaults    SlotDefaults
	slots       []CompiledSlot
	joins       []CompiledJoin
	bgms        []CompiledBGM
	layers      []CompiledLayer
	constraints []ConstraintSpec
}

func (c *CompiledTemplate) ID() ID {
	if c == nil {
		return ""
	}
	return c.id
}

func (c *CompiledTemplate) Fingerprint() string {
	if c == nil {
		return ""
	}
	return c.fingerprint
}

func (c *CompiledTemplate) Canvas() CanvasSpec {
	if c == nil {
		return CanvasSpec{}
	}
	return c.canvas
}

func (c *CompiledTemplate) Background() CompiledBackground {
	if c == nil {
		return CompiledBackground{}
	}
	return cloneCompiledBackground(c.background)
}

func (c *CompiledTemplate) Defaults() SlotDefaults {
	if c == nil {
		return SlotDefaults{}
	}
	return c.defaults
}

func (c *CompiledTemplate) Slots() []CompiledSlot {
	if c == nil {
		return nil
	}
	return cloneCompiledSlots(c.slots)
}

func (c *CompiledTemplate) Joins() []CompiledJoin {
	if c == nil {
		return nil
	}
	return cloneCompiledJoins(c.joins)
}

func (c *CompiledTemplate) BGMs() []CompiledBGM {
	if c == nil {
		return nil
	}
	return append([]CompiledBGM(nil), c.bgms...)
}

func (c *CompiledTemplate) Layers() []CompiledLayer {
	if c == nil {
		return nil
	}
	return cloneCompiledLayers(c.layers)
}

func (c *CompiledTemplate) Constraints() []ConstraintSpec {
	if c == nil {
		return nil
	}
	return cloneConstraints(c.constraints)
}

func cloneCompiledSlots(values []CompiledSlot) []CompiledSlot {
	cloned := make([]CompiledSlot, len(values))
	for index, value := range values {
		cloned[index] = value
		cloned[index].Videos = append([]CompiledVideo(nil), value.Videos...)
	}
	return cloned
}

func cloneCompiledJoins(values []CompiledJoin) []CompiledJoin {
	cloned := make([]CompiledJoin, len(values))
	for index, value := range values {
		cloned[index] = value
		cloned[index].Transitions = append([]CompiledTransition(nil), value.Transitions...)
		cloned[index].compatibility = make(map[compatibilityKey]bool, len(value.compatibility))
		for key, compatible := range value.compatibility {
			cloned[index].compatibility[key] = compatible
		}
	}
	return cloned
}

func cloneCompiledLayers(values []CompiledLayer) []CompiledLayer {
	cloned := make([]CompiledLayer, len(values))
	for index, value := range values {
		cloned[index] = value
		cloned[index].Timing = cloneLayerTiming(value.Timing)
		if value.Image != nil {
			image := *value.Image
			cloned[index].Image = &image
		}
		if value.Subtitle != nil {
			subtitle := *value.Subtitle
			subtitle.Cues = append([]NormalizedCue(nil), value.Subtitle.Cues...)
			if value.Subtitle.Style.Font != nil {
				font := *value.Subtitle.Style.Font
				subtitle.Style.Font = &font
			}
			cloned[index].Subtitle = &subtitle
		}
	}
	return cloned
}

func cloneCompiledBackground(value CompiledBackground) CompiledBackground {
	cloned := CompiledBackground{Kind: value.Kind}
	if value.Color != nil {
		color := *value.Color
		cloned.Color = &color
	}
	if value.Image != nil {
		image := *value.Image
		cloned.Image = &image
	}
	if value.Blur != nil {
		blur := *value.Blur
		cloned.Blur = &blur
	}
	return cloned
}

func cloneConstraints(values []ConstraintSpec) []ConstraintSpec {
	cloned := make([]ConstraintSpec, len(values))
	for index, value := range values {
		cloned[index] = value
		if value.MaxSimilarity != nil {
			payload := *value.MaxSimilarity
			cloned[index].MaxSimilarity = &payload
		}
		if value.MaxVideoAssetUses != nil {
			payload := *value.MaxVideoAssetUses
			cloned[index].MaxVideoAssetUses = &payload
		}
		if value.MaxBGMUses != nil {
			payload := *value.MaxBGMUses
			cloned[index].MaxBGMUses = &payload
		}
	}
	return cloned
}
