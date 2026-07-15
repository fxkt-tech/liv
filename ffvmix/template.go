// Package ffvmix defines persistent mix templates and their one-time local
// asset compilation.
package ffvmix

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffcut"
	"github.com/google/uuid"
)

const TemplateSchemaVersion = 1

type ID = ffcut.ID

type Template struct {
	SchemaVersion int              `json:"schema_version"`
	ID            ID               `json:"id"`
	Canvas        CanvasSpec       `json:"canvas"`
	Background    BackgroundSpec   `json:"background"`
	Defaults      SlotDefaults     `json:"defaults"`
	Slots         []*Slot          `json:"slots"`
	Joins         []*Join          `json:"joins"`
	BGMs          []*BGM           `json:"bgms,omitempty"`
	Layers        []*LayerSpec     `json:"layers,omitempty"`
	Constraints   []ConstraintSpec `json:"constraints,omitempty"`
}

type TemplateConfig struct {
	Canvas     CanvasSpec
	Background BackgroundSpec
	Defaults   *SlotDefaults
}

type CanvasSpec struct {
	Width     int32           `json:"width"`
	Height    int32           `json:"height"`
	FrameRate ffcut.FrameRate `json:"frame_rate"`
}

type BackgroundKind = ffcut.BackgroundKind

const (
	BackgroundKindColor = ffcut.BackgroundKindColor
	BackgroundKindImage = ffcut.BackgroundKindImage
	BackgroundKindBlur  = ffcut.BackgroundKindBlur
)

type BackgroundSpec struct {
	Kind  BackgroundKind       `json:"kind"`
	Color *ColorBackgroundSpec `json:"color,omitempty"`
	Image *ImageBackgroundSpec `json:"image,omitempty"`
	Blur  *BlurBackgroundSpec  `json:"blur,omitempty"`
}

type ColorBackgroundSpec struct {
	Color string `json:"color"`
}

type ImageBackgroundSpec struct {
	Path string        `json:"path"`
	Fit  ffcut.FitMode `json:"fit"`
}

type BlurBackgroundSpec struct {
	Sigma float64 `json:"sigma"`
}

type OverflowPolicy string

const (
	OverflowSpeedUp OverflowPolicy = "speed_up"
	OverflowTrim    OverflowPolicy = "trim"
	OverflowReject  OverflowPolicy = "reject"
)

type UnderflowPolicy string

const (
	UnderflowSlowDown UnderflowPolicy = "slow_down"
	UnderflowLoop     UnderflowPolicy = "loop"
	UnderflowFreeze   UnderflowPolicy = "freeze"
	UnderflowReject   UnderflowPolicy = "reject"
)

type TrimMode string

const (
	TrimStart  TrimMode = "start"
	TrimCenter TrimMode = "center"
	TrimEnd    TrimMode = "end"
	TrimRandom TrimMode = "random"
)

type SlotDefaults struct {
	Fit             ffcut.FitMode   `json:"fit"`
	AudioGain       float64         `json:"audio_gain"`
	Overflow        OverflowPolicy  `json:"overflow"`
	Underflow       UnderflowPolicy `json:"underflow"`
	Trim            TrimMode        `json:"trim"`
	MinPlaybackRate float64         `json:"min_playback_rate"`
	MaxPlaybackRate float64         `json:"max_playback_rate"`
}

type SlotOverrides struct {
	Fit             *ffcut.FitMode   `json:"fit,omitempty"`
	AudioGain       *float64         `json:"audio_gain,omitempty"`
	Overflow        *OverflowPolicy  `json:"overflow,omitempty"`
	Underflow       *UnderflowPolicy `json:"underflow,omitempty"`
	Trim            *TrimMode        `json:"trim,omitempty"`
	MinPlaybackRate *float64         `json:"min_playback_rate,omitempty"`
	MaxPlaybackRate *float64         `json:"max_playback_rate,omitempty"`
}

type Slot struct {
	ID             ID              `json:"id"`
	Name           string          `json:"name,omitempty"`
	TargetDuration *ffcut.Duration `json:"target_duration,omitempty"`
	Overrides      SlotOverrides   `json:"overrides,omitempty"`
	Videos         []*VideoSource  `json:"videos"`
}

type SlotConfig struct {
	Name           string
	TargetDuration *ffcut.Duration
	Overrides      SlotOverrides
}

type VideoSource struct {
	ID          ID               `json:"id"`
	Path        string           `json:"path"`
	SourceRange *ffcut.TimeRange `json:"source_range,omitempty"`
	Weight      float64          `json:"weight"`
	Fit         *ffcut.FitMode   `json:"fit,omitempty"`
	AudioGain   *float64         `json:"audio_gain,omitempty"`
}

type VideoSourceConfig struct {
	Path        string
	SourceRange *ffcut.TimeRange
	Weight      float64
	Fit         *ffcut.FitMode
	AudioGain   *float64
}

type Join struct {
	ID          ID                     `json:"id"`
	FromSlotID  ID                     `json:"from_slot_id"`
	ToSlotID    ID                     `json:"to_slot_id"`
	Transitions []*TransitionCandidate `json:"transitions"`
}

type JoinConfig struct {
	FromSlotID ID
	ToSlotID   ID
}

type TransitionCandidate struct {
	ID             ID                   `json:"id"`
	Kind           ffcut.TransitionKind `json:"kind"`
	Duration       ffcut.Duration       `json:"duration"`
	AudioCrossfade bool                 `json:"audio_crossfade"`
	Weight         float64              `json:"weight"`
}

type TransitionConfig struct {
	Kind           ffcut.TransitionKind
	Duration       ffcut.Duration
	AudioCrossfade bool
	Weight         float64
}

type BGM struct {
	ID            ID               `json:"id"`
	Path          string           `json:"path"`
	SourceRange   *ffcut.TimeRange `json:"source_range,omitempty"`
	TimelineStart ffcut.Duration   `json:"timeline_start,omitempty"`
	Loop          bool             `json:"loop,omitempty"`
	Gain          float64          `json:"gain"`
	FadeIn        ffcut.Duration   `json:"fade_in,omitempty"`
	FadeOut       ffcut.Duration   `json:"fade_out,omitempty"`
	Weight        float64          `json:"weight"`
}

type BGMConfig struct {
	Path          string
	SourceRange   *ffcut.TimeRange
	TimelineStart ffcut.Duration
	Loop          bool
	Gain          *float64
	FadeIn        ffcut.Duration
	FadeOut       ffcut.Duration
	Weight        float64
}

type ConstraintKind string

const (
	ConstraintMaxSimilarity     ConstraintKind = "max_similarity"
	ConstraintMaxVideoAssetUses ConstraintKind = "max_video_asset_uses"
	ConstraintMaxBGMUses        ConstraintKind = "max_bgm_uses"
)

type ConstraintSpec struct {
	ID                ID                     `json:"id"`
	Kind              ConstraintKind         `json:"kind"`
	MaxSimilarity     *MaxSimilaritySpec     `json:"max_similarity,omitempty"`
	MaxVideoAssetUses *MaxVideoAssetUsesSpec `json:"max_video_asset_uses,omitempty"`
	MaxBGMUses        *MaxBGMUsesSpec        `json:"max_bgm_uses,omitempty"`
}

type MaxSimilaritySpec struct {
	Maximum float64 `json:"maximum"`
}

type MaxVideoAssetUsesSpec struct {
	Maximum int `json:"maximum"`
}

type MaxBGMUsesSpec struct {
	Maximum int `json:"maximum"`
}

func NewTemplate(config TemplateConfig) *Template {
	frameRate := config.Canvas.FrameRate
	if frameRate.Numerator == 0 && frameRate.Denominator == 0 {
		frameRate = ffcut.FrameRate{Numerator: 30, Denominator: 1}
	}
	background := cloneBackgroundSpec(config.Background)
	if background.Kind == "" {
		background = BackgroundSpec{
			Kind:  BackgroundKindColor,
			Color: &ColorBackgroundSpec{Color: "#000000"},
		}
	}
	defaults := DefaultSlotDefaults()
	if config.Defaults != nil {
		defaults = *config.Defaults
	}
	return &Template{
		SchemaVersion: TemplateSchemaVersion,
		ID:            newID("template"),
		Canvas: CanvasSpec{
			Width:     config.Canvas.Width,
			Height:    config.Canvas.Height,
			FrameRate: frameRate,
		},
		Background: background,
		Defaults:   defaults,
		Slots:      make([]*Slot, 0),
		Joins:      make([]*Join, 0),
	}
}

func DefaultSlotDefaults() SlotDefaults {
	return SlotDefaults{
		Fit:             ffcut.FitModeCover,
		AudioGain:       1,
		Overflow:        OverflowReject,
		Underflow:       UnderflowReject,
		Trim:            TrimCenter,
		MinPlaybackRate: 0.5,
		MaxPlaybackRate: 2,
	}
}

func (t *Template) AddSlot(config SlotConfig) (*Slot, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: template is nil", ErrInvalidTemplate)
	}
	slot := &Slot{
		ID:             newID("slot"),
		Name:           config.Name,
		TargetDuration: cloneDuration(config.TargetDuration),
		Overrides:      cloneSlotOverrides(config.Overrides),
		Videos:         make([]*VideoSource, 0),
	}
	t.Slots = append(t.Slots, slot)
	return slot, nil
}

func (s *Slot) AddVideo(config VideoSourceConfig) (*VideoSource, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: slot is nil", ErrInvalidTemplate)
	}
	weight := defaultWeight(config.Weight)
	video := &VideoSource{
		ID:          newID("video"),
		Path:        config.Path,
		SourceRange: cloneRange(config.SourceRange),
		Weight:      weight,
		Fit:         cloneFit(config.Fit),
		AudioGain:   cloneFloat(config.AudioGain),
	}
	s.Videos = append(s.Videos, video)
	return video, nil
}

func (t *Template) AddJoin(config JoinConfig) (*Join, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: template is nil", ErrInvalidTemplate)
	}
	join := &Join{
		ID:          newID("join"),
		FromSlotID:  config.FromSlotID,
		ToSlotID:    config.ToSlotID,
		Transitions: make([]*TransitionCandidate, 0),
	}
	t.Joins = append(t.Joins, join)
	return join, nil
}

func (j *Join) AddTransition(config TransitionConfig) (*TransitionCandidate, error) {
	if j == nil {
		return nil, fmt.Errorf("%w: join is nil", ErrInvalidTemplate)
	}
	transition := &TransitionCandidate{
		ID:             newID("transition"),
		Kind:           config.Kind,
		Duration:       config.Duration,
		AudioCrossfade: config.AudioCrossfade,
		Weight:         defaultWeight(config.Weight),
	}
	j.Transitions = append(j.Transitions, transition)
	return transition, nil
}

func (t *Template) AddBGM(config BGMConfig) (*BGM, error) {
	if t == nil {
		return nil, fmt.Errorf("%w: template is nil", ErrInvalidTemplate)
	}
	gain := 1.0
	if config.Gain != nil {
		gain = *config.Gain
	}
	bgm := &BGM{
		ID:            newID("bgm"),
		Path:          config.Path,
		SourceRange:   cloneRange(config.SourceRange),
		TimelineStart: config.TimelineStart,
		Loop:          config.Loop,
		Gain:          gain,
		FadeIn:        config.FadeIn,
		FadeOut:       config.FadeOut,
		Weight:        defaultWeight(config.Weight),
	}
	t.BGMs = append(t.BGMs, bgm)
	return bgm, nil
}

func (t *Template) AddMaxSimilarity(maximum float64) ConstraintSpec {
	spec := ConstraintSpec{
		ID:            newID("constraint"),
		Kind:          ConstraintMaxSimilarity,
		MaxSimilarity: &MaxSimilaritySpec{Maximum: maximum},
	}
	t.Constraints = append(t.Constraints, spec)
	return spec
}

func (t *Template) AddMaxVideoAssetUses(maximum int) ConstraintSpec {
	spec := ConstraintSpec{
		ID:                newID("constraint"),
		Kind:              ConstraintMaxVideoAssetUses,
		MaxVideoAssetUses: &MaxVideoAssetUsesSpec{Maximum: maximum},
	}
	t.Constraints = append(t.Constraints, spec)
	return spec
}

func (t *Template) AddMaxBGMUses(maximum int) ConstraintSpec {
	spec := ConstraintSpec{
		ID:         newID("constraint"),
		Kind:       ConstraintMaxBGMUses,
		MaxBGMUses: &MaxBGMUsesSpec{Maximum: maximum},
	}
	t.Constraints = append(t.Constraints, spec)
	return spec
}

func newID(prefix string) ID {
	return ID(prefix + "-" + uuid.NewString())
}

func defaultWeight(weight float64) float64 {
	if weight == 0 {
		return 1
	}
	return weight
}

func cloneDuration(value *ffcut.Duration) *ffcut.Duration {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneRange(value *ffcut.TimeRange) *ffcut.TimeRange {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneFit(value *ffcut.FitMode) *ffcut.FitMode {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneSlotOverrides(value SlotOverrides) SlotOverrides {
	return SlotOverrides{
		Fit:             cloneFit(value.Fit),
		AudioGain:       cloneFloat(value.AudioGain),
		Overflow:        cloneOverflow(value.Overflow),
		Underflow:       cloneUnderflow(value.Underflow),
		Trim:            cloneTrim(value.Trim),
		MinPlaybackRate: cloneFloat(value.MinPlaybackRate),
		MaxPlaybackRate: cloneFloat(value.MaxPlaybackRate),
	}
}

func cloneBackgroundSpec(value BackgroundSpec) BackgroundSpec {
	cloned := BackgroundSpec{Kind: value.Kind}
	if value.Color != nil {
		payload := *value.Color
		cloned.Color = &payload
	}
	if value.Image != nil {
		payload := *value.Image
		cloned.Image = &payload
	}
	if value.Blur != nil {
		payload := *value.Blur
		cloned.Blur = &payload
	}
	return cloned
}

func cloneOverflow(value *OverflowPolicy) *OverflowPolicy {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneUnderflow(value *UnderflowPolicy) *UnderflowPolicy {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneTrim(value *TrimMode) *TrimMode {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
