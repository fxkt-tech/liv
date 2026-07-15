package ffvmix

import (
	"fmt"
	"math"
	"strings"

	"github.com/fxkt-tech/liv/ffcut"
)

func (t *Template) Validate() error {
	validator := newTemplateValidator()
	validator.validateTemplate(t)
	return validator.err()
}

type templateValidator struct {
	issues  []Issue
	ids     map[ID]string
	slotIDs map[ID]string
}

func newTemplateValidator() *templateValidator {
	return &templateValidator{
		ids:     make(map[ID]string),
		slotIDs: make(map[ID]string),
	}
}

func (v *templateValidator) add(code IssueCode, path, message string, cause error) {
	if cause == nil {
		cause = ErrInvalidTemplate
	}
	v.issues = append(v.issues, Issue{Code: code, Path: path, Message: message, Cause: cause})
}

func (v *templateValidator) err() error {
	if len(v.issues) == 0 {
		return nil
	}
	return &CompileError{Issues: append([]Issue(nil), v.issues...)}
}

func (v *templateValidator) validateTemplate(template *Template) {
	if template == nil {
		v.add(IssueInvalidValue, "template", "is required", nil)
		return
	}
	if template.SchemaVersion != TemplateSchemaVersion {
		v.add(IssueInvalidValue, "schema_version", fmt.Sprintf("must be %d, got %d", TemplateSchemaVersion, template.SchemaVersion), nil)
	}
	v.addID("id", template.ID)
	v.validateCanvas("canvas", template.Canvas)
	v.validateBackground("background", template.Background)
	v.validateDefaults("defaults", template.Defaults)

	if len(template.Slots) == 0 {
		v.add(IssueInvalidValue, "slots", "must contain at least one slot", nil)
	}
	for index, slot := range template.Slots {
		path := fmt.Sprintf("slots[%d]", index)
		if slot == nil {
			v.add(IssueInvalidValue, path, "must not be null", nil)
			continue
		}
		v.addID(path+".id", slot.ID)
		if slot.ID != "" {
			v.slotIDs[slot.ID] = path
		}
	}
	for index, slot := range template.Slots {
		if slot != nil {
			v.validateSlot(fmt.Sprintf("slots[%d]", index), *slot, template.Defaults)
		}
	}
	v.validateJoins("joins", template.Joins, template.Slots)
	for index, bgm := range template.BGMs {
		path := fmt.Sprintf("bgms[%d]", index)
		if bgm == nil {
			v.add(IssueInvalidValue, path, "must not be null", nil)
			continue
		}
		v.validateBGM(path, *bgm)
	}
	for index, layer := range template.Layers {
		path := fmt.Sprintf("layers[%d]", index)
		if layer == nil {
			v.add(IssueInvalidValue, path, "must not be null", nil)
			continue
		}
		v.validateLayer(path, *layer)
	}
	for index, constraint := range template.Constraints {
		v.validateConstraint(fmt.Sprintf("constraints[%d]", index), constraint)
	}
}

func (v *templateValidator) addID(path string, id ID) {
	if strings.TrimSpace(string(id)) == "" {
		v.add(IssueInvalidID, path, "must not be empty", nil)
		return
	}
	if previous, exists := v.ids[id]; exists {
		v.add(IssueInvalidID, path, fmt.Sprintf("duplicates ID declared at %s", previous), nil)
		return
	}
	v.ids[id] = path
}

func (v *templateValidator) validateCanvas(path string, canvas CanvasSpec) {
	if canvas.Width <= 0 {
		v.add(IssueInvalidValue, path+".width", "must be positive", nil)
	}
	if canvas.Height <= 0 {
		v.add(IssueInvalidValue, path+".height", "must be positive", nil)
	}
	if canvas.FrameRate.Numerator <= 0 {
		v.add(IssueInvalidValue, path+".frame_rate.numerator", "must be positive", nil)
	}
	if canvas.FrameRate.Denominator <= 0 {
		v.add(IssueInvalidValue, path+".frame_rate.denominator", "must be positive", nil)
	}
}

func (v *templateValidator) validateBackground(path string, background BackgroundSpec) {
	payloads := countPresent(background.Color != nil, background.Image != nil, background.Blur != nil)
	if payloads != 1 {
		v.add(IssueInvalidUnion, path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}
	switch background.Kind {
	case BackgroundKindColor:
		if background.Color == nil {
			v.add(IssueInvalidUnion, path+".color", "is required when kind is color", nil)
		}
	case BackgroundKindImage:
		if background.Image == nil {
			v.add(IssueInvalidUnion, path+".image", "is required when kind is image", nil)
		}
	case BackgroundKindBlur:
		if background.Blur == nil {
			v.add(IssueInvalidUnion, path+".blur", "is required when kind is blur", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".kind", fmt.Sprintf("has unsupported value %q", background.Kind), nil)
	}
	if background.Color != nil && strings.TrimSpace(background.Color.Color) == "" {
		v.add(IssueInvalidValue, path+".color.color", "must not be empty", nil)
	}
	if background.Image != nil {
		v.requirePath(path+".image.path", background.Image.Path)
		v.validateFit(path+".image.fit", background.Image.Fit)
	}
	if background.Blur != nil && (!finite(background.Blur.Sigma) || background.Blur.Sigma <= 0) {
		v.add(IssueInvalidValue, path+".blur.sigma", "must be finite and positive", nil)
	}
}

func (v *templateValidator) validateDefaults(path string, defaults SlotDefaults) {
	v.validateFit(path+".fit", defaults.Fit)
	v.validateGain(path+".audio_gain", defaults.AudioGain)
	v.validateOverflow(path+".overflow", defaults.Overflow)
	v.validateUnderflow(path+".underflow", defaults.Underflow)
	v.validateTrim(path+".trim", defaults.Trim)
	v.validateRates(path, defaults.MinPlaybackRate, defaults.MaxPlaybackRate)
}

func (v *templateValidator) validateSlot(path string, slot Slot, defaults SlotDefaults) {
	if slot.TargetDuration != nil {
		v.validateDuration(path+".target_duration", *slot.TargetDuration, false)
	}
	v.validateOverrides(path+".overrides", slot.Overrides, defaults)
	if len(slot.Videos) == 0 {
		v.add(IssueInvalidValue, path+".videos", "must contain at least one video", nil)
	}
	for index, video := range slot.Videos {
		videoPath := fmt.Sprintf("%s.videos[%d]", path, index)
		if video == nil {
			v.add(IssueInvalidValue, videoPath, "must not be null", nil)
			continue
		}
		v.addID(videoPath+".id", video.ID)
		v.requirePath(videoPath+".path", video.Path)
		if video.SourceRange != nil {
			v.validateRange(videoPath+".source_range", *video.SourceRange)
		}
		v.validateWeight(videoPath+".weight", video.Weight)
		if video.Fit != nil {
			v.validateFit(videoPath+".fit", *video.Fit)
		}
		if video.AudioGain != nil {
			v.validateGain(videoPath+".audio_gain", *video.AudioGain)
		}
	}
}

func (v *templateValidator) validateOverrides(path string, overrides SlotOverrides, defaults SlotDefaults) {
	if overrides.Fit != nil {
		v.validateFit(path+".fit", *overrides.Fit)
	}
	if overrides.AudioGain != nil {
		v.validateGain(path+".audio_gain", *overrides.AudioGain)
	}
	if overrides.Overflow != nil {
		v.validateOverflow(path+".overflow", *overrides.Overflow)
	}
	if overrides.Underflow != nil {
		v.validateUnderflow(path+".underflow", *overrides.Underflow)
	}
	if overrides.Trim != nil {
		v.validateTrim(path+".trim", *overrides.Trim)
	}
	minimum := defaults.MinPlaybackRate
	maximum := defaults.MaxPlaybackRate
	ratesOverridden := false
	if overrides.MinPlaybackRate != nil {
		minimum = *overrides.MinPlaybackRate
		ratesOverridden = true
	}
	if overrides.MaxPlaybackRate != nil {
		maximum = *overrides.MaxPlaybackRate
		ratesOverridden = true
	}
	if ratesOverridden {
		v.validateRates(path, minimum, maximum)
	}
}

func (v *templateValidator) validateJoins(path string, joins []*Join, slots []*Slot) {
	expected := 0
	if len(slots) > 0 {
		expected = len(slots) - 1
	}
	if len(joins) != expected {
		v.add(IssueInvalidValue, path, fmt.Sprintf("must contain %d joins, got %d", expected, len(joins)), nil)
	}
	for index, join := range joins {
		joinPath := fmt.Sprintf("%s[%d]", path, index)
		if join == nil {
			v.add(IssueInvalidValue, joinPath, "must not be null", nil)
			continue
		}
		v.addID(joinPath+".id", join.ID)
		if index < expected && slots[index] != nil && slots[index+1] != nil {
			if join.FromSlotID != slots[index].ID {
				v.add(IssueInvalidReference, joinPath+".from_slot_id", fmt.Sprintf("must reference adjacent slot %q", slots[index].ID), nil)
			}
			if join.ToSlotID != slots[index+1].ID {
				v.add(IssueInvalidReference, joinPath+".to_slot_id", fmt.Sprintf("must reference adjacent slot %q", slots[index+1].ID), nil)
			}
		} else {
			v.validateSlotReference(joinPath+".from_slot_id", join.FromSlotID)
			v.validateSlotReference(joinPath+".to_slot_id", join.ToSlotID)
		}
		if len(join.Transitions) == 0 {
			v.add(IssueInvalidValue, joinPath+".transitions", "must contain at least one transition", nil)
		}
		for transitionIndex, transition := range join.Transitions {
			transitionPath := fmt.Sprintf("%s.transitions[%d]", joinPath, transitionIndex)
			if transition == nil {
				v.add(IssueInvalidValue, transitionPath, "must not be null", nil)
				continue
			}
			v.addID(transitionPath+".id", transition.ID)
			v.validateTransition(transitionPath, *transition)
		}
	}
}

func (v *templateValidator) validateTransition(path string, transition TransitionCandidate) {
	v.validateWeight(path+".weight", transition.Weight)
	switch transition.Kind {
	case ffcut.TransitionKindCut:
		v.validateDuration(path+".duration", transition.Duration, true)
		if transition.Duration != 0 {
			v.add(IssueInvalidValue, path+".duration", "must be zero for a cut", nil)
		}
		if transition.AudioCrossfade {
			v.add(IssueInvalidValue, path+".audio_crossfade", "must be false for a cut", nil)
		}
	case ffcut.TransitionKindFade:
		v.validateDuration(path+".duration", transition.Duration, false)
	default:
		v.add(IssueInvalidValue, path+".kind", fmt.Sprintf("has unsupported value %q", transition.Kind), nil)
	}
}

func (v *templateValidator) validateBGM(path string, bgm BGM) {
	v.addID(path+".id", bgm.ID)
	v.requirePath(path+".path", bgm.Path)
	if bgm.SourceRange != nil {
		v.validateRange(path+".source_range", *bgm.SourceRange)
	}
	v.validateDuration(path+".timeline_start", bgm.TimelineStart, true)
	v.validateDuration(path+".fade_in", bgm.FadeIn, true)
	v.validateDuration(path+".fade_out", bgm.FadeOut, true)
	v.validateGain(path+".gain", bgm.Gain)
	v.validateWeight(path+".weight", bgm.Weight)
}

func (v *templateValidator) validateLayer(path string, layer LayerSpec) {
	v.addID(path+".id", layer.ID)
	v.validateLayerTiming(path+".timing", layer.Timing)
	payloads := countPresent(layer.Image != nil, layer.Subtitle != nil)
	if payloads != 1 {
		v.add(IssueInvalidUnion, path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}
	switch layer.Kind {
	case LayerKindImage:
		if layer.Image == nil {
			v.add(IssueInvalidUnion, path+".image", "is required when kind is image", nil)
		}
	case LayerKindSubtitle:
		if layer.Subtitle == nil {
			v.add(IssueInvalidUnion, path+".subtitle", "is required when kind is subtitle", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".kind", fmt.Sprintf("has unsupported value %q", layer.Kind), nil)
	}
	if layer.Image != nil {
		v.requirePath(path+".image.path", layer.Image.Path)
		v.validateGeometry(path+".image.geometry", layer.Image.Geometry)
		if !finite(layer.Image.Opacity) || layer.Image.Opacity < 0 || layer.Image.Opacity > 1 {
			v.add(IssueInvalidValue, path+".image.opacity", "must be between 0 and 1", nil)
		}
		if !finite(layer.Image.RotationDegrees) {
			v.add(IssueInvalidValue, path+".image.rotation_degrees", "must be finite", nil)
		}
	}
	if layer.Subtitle != nil {
		v.validateSubtitleLayer(path+".subtitle", *layer.Subtitle)
	}
}

func (v *templateValidator) validateLayerTiming(path string, timing LayerTiming) {
	payloads := countPresent(timing.FullOutput != nil, timing.Slot != nil, timing.Absolute != nil)
	if payloads != 1 {
		v.add(IssueInvalidUnion, path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}
	switch timing.Kind {
	case LayerTimeFullOutput:
		if timing.FullOutput == nil {
			v.add(IssueInvalidUnion, path+".full_output", "is required for full_output timing", nil)
		}
	case LayerTimeSlot:
		if timing.Slot == nil {
			v.add(IssueInvalidUnion, path+".slot", "is required for slot timing", nil)
		}
	case LayerTimeAbsolute:
		if timing.Absolute == nil {
			v.add(IssueInvalidUnion, path+".absolute", "is required for absolute timing", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".kind", fmt.Sprintf("has unsupported value %q", timing.Kind), nil)
	}
	if timing.Slot != nil {
		v.validateSlotReference(path+".slot.slot_id", timing.Slot.SlotID)
		v.validateDuration(path+".slot.offset", timing.Slot.Offset, true)
		if timing.Slot.Duration != nil {
			v.validateDuration(path+".slot.duration", *timing.Slot.Duration, false)
			if _, err := (ffcut.TimeRange{Start: timing.Slot.Offset, Duration: *timing.Slot.Duration}).End(); err != nil {
				v.add(IssueInvalidValue, path+".slot", "has invalid end", err)
			}
		}
	}
	if timing.Absolute != nil {
		v.validateRange(path+".absolute.range", timing.Absolute.Range)
	}
}

func (v *templateValidator) validateSubtitleLayer(path string, subtitle SubtitleLayerSpec) {
	v.validateGeometry(path+".region", subtitle.Region)
	v.validateSubtitleStyle(path+".style", subtitle.Style)
	payloads := countPresent(subtitle.Input.Structured != nil, subtitle.Input.SRT != nil, subtitle.Input.ASS != nil)
	if payloads != 1 {
		v.add(IssueInvalidUnion, path+".input", fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}
	switch subtitle.Input.Kind {
	case SubtitleInputStructured:
		if subtitle.Input.Structured == nil {
			v.add(IssueInvalidUnion, path+".input.structured", "is required for structured subtitles", nil)
		}
	case SubtitleInputSRT:
		if subtitle.Input.SRT == nil {
			v.add(IssueInvalidUnion, path+".input.srt", "is required for SRT subtitles", nil)
		}
	case SubtitleInputASS:
		if subtitle.Input.ASS == nil {
			v.add(IssueInvalidUnion, path+".input.ass", "is required for ASS subtitles", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".input.kind", fmt.Sprintf("has unsupported value %q", subtitle.Input.Kind), nil)
	}
	if subtitle.Input.Structured != nil {
		if len(subtitle.Input.Structured.Cues) == 0 {
			v.add(IssueInvalidValue, path+".input.structured.cues", "must contain at least one cue", nil)
		}
		for index, cue := range subtitle.Input.Structured.Cues {
			cuePath := fmt.Sprintf("%s.input.structured.cues[%d]", path, index)
			v.addID(cuePath+".id", cue.ID)
			v.validateRange(cuePath+".range", cue.Range)
			if strings.TrimSpace(cue.Text) == "" {
				v.add(IssueInvalidValue, cuePath+".text", "must not be empty", nil)
			}
		}
	}
	if subtitle.Input.SRT != nil {
		v.requirePath(path+".input.srt.path", subtitle.Input.SRT.Path)
	}
	if subtitle.Input.ASS != nil {
		v.requirePath(path+".input.ass.path", subtitle.Input.ASS.Path)
	}
}

func (v *templateValidator) validateSubtitleStyle(path string, style SubtitleStyleSpec) {
	if strings.TrimSpace(style.FontFamily) == "" && strings.TrimSpace(style.FontPath) == "" {
		v.add(IssueInvalidValue, path+".font_family", "or font_path must be provided", nil)
	}
	if style.FontPath != "" {
		v.requirePath(path+".font_path", style.FontPath)
	}
	v.validateLength(path+".font_size", style.FontSize, false, true)
	if strings.TrimSpace(style.Color) == "" {
		v.add(IssueInvalidValue, path+".color", "must not be empty", nil)
	}
	switch style.Align {
	case ffcut.TextAlignLeft, ffcut.TextAlignCenter, ffcut.TextAlignRight:
	default:
		v.add(IssueInvalidValue, path+".align", fmt.Sprintf("has unsupported value %q", style.Align), nil)
	}
}

func (v *templateValidator) validateConstraint(path string, constraint ConstraintSpec) {
	v.addID(path+".id", constraint.ID)
	payloads := countPresent(constraint.MaxSimilarity != nil, constraint.MaxVideoAssetUses != nil, constraint.MaxBGMUses != nil)
	if payloads != 1 {
		v.add(IssueInvalidUnion, path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}
	switch constraint.Kind {
	case ConstraintMaxSimilarity:
		if constraint.MaxSimilarity == nil {
			v.add(IssueInvalidUnion, path+".max_similarity", "is required for max_similarity", nil)
		}
	case ConstraintMaxVideoAssetUses:
		if constraint.MaxVideoAssetUses == nil {
			v.add(IssueInvalidUnion, path+".max_video_asset_uses", "is required for max_video_asset_uses", nil)
		}
	case ConstraintMaxBGMUses:
		if constraint.MaxBGMUses == nil {
			v.add(IssueInvalidUnion, path+".max_bgm_uses", "is required for max_bgm_uses", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".kind", fmt.Sprintf("has unsupported value %q", constraint.Kind), nil)
	}
	if constraint.MaxSimilarity != nil && (!finite(constraint.MaxSimilarity.Maximum) || constraint.MaxSimilarity.Maximum < 0 || constraint.MaxSimilarity.Maximum > 1) {
		v.add(IssueInvalidValue, path+".max_similarity.maximum", "must be between 0 and 1", nil)
	}
	if constraint.MaxVideoAssetUses != nil && constraint.MaxVideoAssetUses.Maximum <= 0 {
		v.add(IssueInvalidValue, path+".max_video_asset_uses.maximum", "must be positive", nil)
	}
	if constraint.MaxBGMUses != nil && constraint.MaxBGMUses.Maximum <= 0 {
		v.add(IssueInvalidValue, path+".max_bgm_uses.maximum", "must be positive", nil)
	}
}

func (v *templateValidator) validateSlotReference(path string, id ID) {
	if strings.TrimSpace(string(id)) == "" {
		v.add(IssueInvalidReference, path, "must not be empty", nil)
		return
	}
	if _, exists := v.slotIDs[id]; !exists {
		v.add(IssueInvalidReference, path, fmt.Sprintf("references unknown slot %q", id), nil)
	}
}

func (v *templateValidator) validateRange(path string, value ffcut.TimeRange) {
	v.validateDuration(path+".start", value.Start, true)
	v.validateDuration(path+".duration", value.Duration, false)
	if _, err := value.End(); err != nil {
		v.add(IssueInvalidValue, path, "has invalid end", err)
	}
}

func (v *templateValidator) validateDuration(path string, value ffcut.Duration, allowZero bool) {
	if value < 0 {
		v.add(IssueInvalidValue, path, "must be non-negative", ffcut.ErrInvalidDuration)
		return
	}
	if !allowZero && value == 0 {
		v.add(IssueInvalidValue, path, "must be positive", nil)
		return
	}
	if _, err := value.Microseconds(); err != nil {
		v.add(IssueInvalidValue, path, "must use microsecond precision", err)
	}
}

func (v *templateValidator) validateRates(path string, minimum, maximum float64) {
	if !finite(minimum) || minimum <= 0 || minimum > 1 {
		v.add(IssueInvalidValue, path+".min_playback_rate", "must be finite, positive and at most 1", nil)
	}
	if !finite(maximum) || maximum < 1 {
		v.add(IssueInvalidValue, path+".max_playback_rate", "must be finite and at least 1", nil)
	}
	if finite(minimum) && finite(maximum) && minimum > maximum {
		v.add(IssueInvalidValue, path, "minimum playback rate must not exceed maximum", nil)
	}
}

func (v *templateValidator) validateOverflow(path string, value OverflowPolicy) {
	switch value {
	case OverflowSpeedUp, OverflowTrim, OverflowReject:
	default:
		v.add(IssueInvalidValue, path, fmt.Sprintf("has unsupported value %q", value), nil)
	}
}

func (v *templateValidator) validateUnderflow(path string, value UnderflowPolicy) {
	switch value {
	case UnderflowSlowDown, UnderflowLoop, UnderflowFreeze, UnderflowReject:
	default:
		v.add(IssueInvalidValue, path, fmt.Sprintf("has unsupported value %q", value), nil)
	}
}

func (v *templateValidator) validateTrim(path string, value TrimMode) {
	switch value {
	case TrimStart, TrimCenter, TrimEnd, TrimRandom:
	default:
		v.add(IssueInvalidValue, path, fmt.Sprintf("has unsupported value %q", value), nil)
	}
}

func (v *templateValidator) validateFit(path string, value ffcut.FitMode) {
	switch value {
	case ffcut.FitModeCover, ffcut.FitModeContain, ffcut.FitModeStretch:
	default:
		v.add(IssueInvalidValue, path, fmt.Sprintf("has unsupported value %q", value), nil)
	}
}

func (v *templateValidator) validateGain(path string, value float64) {
	if !finite(value) || value < 0 {
		v.add(IssueInvalidValue, path, "must be finite and non-negative", nil)
	}
}

func (v *templateValidator) validateWeight(path string, value float64) {
	if !finite(value) || value <= 0 {
		v.add(IssueInvalidValue, path, "must be finite and positive", nil)
	}
}

func (v *templateValidator) validateGeometry(path string, geometry ffcut.Geometry) {
	switch geometry.Anchor {
	case ffcut.AnchorTopLeft, ffcut.AnchorTopRight, ffcut.AnchorBottomLeft, ffcut.AnchorBottomRight, ffcut.AnchorCenter:
	default:
		v.add(IssueInvalidValue, path+".anchor", fmt.Sprintf("has unsupported value %q", geometry.Anchor), nil)
	}
	v.validateLength(path+".x", geometry.X, true, false)
	v.validateLength(path+".y", geometry.Y, true, false)
	v.validateLength(path+".width", geometry.Width, false, true)
	v.validateLength(path+".height", geometry.Height, false, true)
}

func (v *templateValidator) validateLength(path string, length ffcut.Length, allowNegative, requirePositive bool) {
	if !finite(length.Value) {
		v.add(IssueInvalidValue, path+".value", "must be finite", nil)
		return
	}
	switch length.Unit {
	case ffcut.LengthUnitPixel:
		if !allowNegative && length.Value < 0 {
			v.add(IssueInvalidValue, path+".value", "must be non-negative", nil)
		}
	case ffcut.LengthUnitPercent:
		if length.Value < 0 || length.Value > 100 {
			v.add(IssueInvalidValue, path+".value", "percentage must be between 0 and 100", nil)
		}
	default:
		v.add(IssueInvalidValue, path+".unit", fmt.Sprintf("has unsupported value %q", length.Unit), nil)
	}
	if requirePositive && length.Value <= 0 {
		v.add(IssueInvalidValue, path+".value", "must be positive", nil)
	}
}

func (v *templateValidator) requirePath(path, value string) {
	if strings.TrimSpace(value) == "" {
		v.add(IssueInvalidValue, path, "must not be empty", nil)
	}
}

func countPresent(values ...bool) int {
	count := 0
	for _, present := range values {
		if present {
			count++
		}
	}
	return count
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
