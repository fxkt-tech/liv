package ffcut

import (
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"
)

func (p *Project) Validate() error {
	validator := newProjectValidator()
	if p == nil {
		validator.add("project", "is required", nil)
		return validator.err()
	}

	if p.Version != ProjectVersion {
		validator.add("version", fmt.Sprintf("must be %d, got %d", ProjectVersion, p.Version), ErrUnsupportedVersion)
	}
	validator.addObjectID("id", p.ID)
	validator.validateCanvas("canvas", p.Canvas)
	projectDuration := validator.validateSequence("video", p.Video)
	for index, track := range p.Audio {
		validator.validateAudioTrack(fmt.Sprintf("audio[%d]", index), track, projectDuration)
	}
	for index, layer := range p.Layers {
		validator.validateLayer(fmt.Sprintf("layers[%d]", index), layer, projectDuration)
	}
	validator.validateMetadata("metadata", p.Metadata)

	return validator.err()
}

type projectValidator struct {
	issues []ValidationIssue
	ids    map[ID]string
}

func newProjectValidator() *projectValidator {
	return &projectValidator{ids: make(map[ID]string)}
}

func (v *projectValidator) add(path, message string, cause error) {
	v.issues = append(v.issues, ValidationIssue{Path: path, Message: message, Cause: cause})
}

func (v *projectValidator) err() error {
	if len(v.issues) == 0 {
		return nil
	}
	return &ValidationError{Issues: append([]ValidationIssue(nil), v.issues...)}
}

func (v *projectValidator) addObjectID(path string, id ID) {
	if strings.TrimSpace(string(id)) == "" {
		v.add(path, "must not be empty", nil)
		return
	}
	if previous, exists := v.ids[id]; exists {
		v.add(path, fmt.Sprintf("duplicates ID declared at %s", previous), nil)
		return
	}
	v.ids[id] = path
}

func (v *projectValidator) validateCanvas(path string, canvas Canvas) {
	if canvas.Width <= 0 {
		v.add(path+".width", "must be positive", nil)
	}
	if canvas.Height <= 0 {
		v.add(path+".height", "must be positive", nil)
	}
	if canvas.FrameRate.Numerator <= 0 {
		v.add(path+".frame_rate.numerator", "must be positive", nil)
	}
	if canvas.FrameRate.Denominator <= 0 {
		v.add(path+".frame_rate.denominator", "must be positive", nil)
	}
	v.validateBackground(path+".background", canvas.Background)
}

func (v *projectValidator) validateBackground(path string, background Background) {
	payloads := 0
	if background.Color != nil {
		payloads++
	}
	if background.Image != nil {
		payloads++
	}
	if background.Blur != nil {
		payloads++
	}
	if payloads != 1 {
		v.add(path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}

	switch background.Kind {
	case BackgroundKindColor:
		if background.Color == nil {
			v.add(path+".color", "is required when kind is color", nil)
		}
	case BackgroundKindImage:
		if background.Image == nil {
			v.add(path+".image", "is required when kind is image", nil)
		}
	case BackgroundKindBlur:
		if background.Blur == nil {
			v.add(path+".blur", "is required when kind is blur", nil)
		}
	default:
		v.add(path+".kind", fmt.Sprintf("has unsupported value %q", background.Kind), nil)
	}

	if background.Color != nil && strings.TrimSpace(background.Color.Color) == "" {
		v.add(path+".color.color", "must not be empty", nil)
	}
	if background.Image != nil {
		v.validateLocalSource(path+".image.source", background.Image.Source)
		v.validateFitMode(path+".image.fit", background.Image.Fit)
	}
	if background.Blur != nil && (!finite(background.Blur.Sigma) || background.Blur.Sigma <= 0) {
		v.add(path+".blur.sigma", "must be finite and positive", nil)
	}
}

func (v *projectValidator) validateSequence(path string, sequence Sequence) Duration {
	if len(sequence.Clips) == 0 {
		v.add(path+".clips", "must contain at least one clip", nil)
		return 0
	}

	clipEnds := make([]Duration, len(sequence.Clips))
	clipRangesValid := make([]bool, len(sequence.Clips))
	for index, clip := range sequence.Clips {
		clipPath := fmt.Sprintf("%s.clips[%d]", path, index)
		v.addObjectID(clipPath+".id", clip.ID)
		v.validateLocalSource(clipPath+".source", clip.Source)
		_, sourceRangeValid := v.validateRange(clipPath+".source_range", clip.SourceRange, false)
		end, ok := v.validateRange(clipPath+".timeline_range", clip.TimelineRange, false)
		clipEnds[index], clipRangesValid[index] = end, ok
		if index == 0 && clip.TimelineRange.Start != 0 {
			v.add(clipPath+".timeline_range.start", "first clip must start at zero", nil)
		}
		v.validatePlayback(
			clipPath+".playback",
			clip.Playback,
			clip.SourceRange.Duration,
			clip.TimelineRange.Duration,
			sourceRangeValid && ok,
		)
		v.validateFitMode(clipPath+".fit", clip.Fit)
		v.validateClipAudio(clipPath+".audio", clip.Audio)
	}

	expectedTransitions := len(sequence.Clips) - 1
	if len(sequence.Transitions) != expectedTransitions {
		v.add(path+".transitions", fmt.Sprintf("must contain %d transitions, got %d", expectedTransitions, len(sequence.Transitions)), nil)
	}
	for index, transition := range sequence.Transitions {
		transitionPath := fmt.Sprintf("%s.transitions[%d]", path, index)
		v.addObjectID(transitionPath+".id", transition.ID)
		_, transitionRangeValid := v.validateRange(transitionPath+".range", transition.Range, true)
		if index >= expectedTransitions {
			continue
		}

		from := sequence.Clips[index]
		to := sequence.Clips[index+1]
		if transition.FromClipID != from.ID {
			v.add(transitionPath+".from_clip_id", fmt.Sprintf("must reference adjacent clip %q", from.ID), nil)
		}
		if transition.ToClipID != to.ID {
			v.add(transitionPath+".to_clip_id", fmt.Sprintf("must reference adjacent clip %q", to.ID), nil)
		}
		if !transitionRangeValid || !clipRangesValid[index] || !clipRangesValid[index+1] {
			continue
		}
		v.validateTransitionTiming(transitionPath, transition, from, to, clipEnds[index])
	}

	if clipRangesValid[len(clipRangesValid)-1] {
		return clipEnds[len(clipEnds)-1]
	}
	return 0
}

func (v *projectValidator) validateTransitionTiming(path string, transition Transition, from, to VideoClip, fromEnd Duration) {
	transitionEnd, err := transition.Range.End()
	if err != nil {
		v.add(path+".range", "has invalid end", err)
		return
	}

	switch transition.Kind {
	case TransitionKindCut:
		if transition.Range.Duration != 0 {
			v.add(path+".range.duration", "must be zero for a cut", nil)
		}
		if transition.Range.Start != fromEnd || to.TimelineRange.Start != fromEnd {
			v.add(path+".range", "cut must be located at the exact boundary between adjacent clips", nil)
		}
		if transition.AudioCrossfade {
			v.add(path+".audio_crossfade", "must be false for a cut", nil)
		}
	case TransitionKindFade:
		if transition.Range.Duration <= 0 {
			v.add(path+".range.duration", "must be positive for a fade", nil)
		}
		if transition.Range.Duration > from.TimelineRange.Duration || transition.Range.Duration > to.TimelineRange.Duration {
			v.add(path+".range.duration", "must not exceed either adjacent clip duration", nil)
		}
		if transition.Range.Start != to.TimelineRange.Start || transitionEnd != fromEnd {
			v.add(path+".range", "fade range must equal the overlap between adjacent clips", nil)
		}
	default:
		v.add(path+".kind", fmt.Sprintf("has unsupported value %q", transition.Kind), nil)
	}
}

func (v *projectValidator) validatePlayback(
	path string,
	playback Playback,
	sourceDuration Duration,
	timelineDuration Duration,
	rangesValid bool,
) {
	rateValid := finite(playback.Rate) && playback.Rate > 0
	if !rateValid {
		v.add(path+".rate", "must be finite and positive", nil)
	}
	freezeValid := v.validateDuration(path+".freeze_last_frame", playback.FreezeLastFrame, true)
	if !freezeValid {
		return
	}
	if playback.FreezeLastFrame > timelineDuration {
		v.add(path+".freeze_last_frame", "must not exceed timeline duration", nil)
	}
	if playback.Loop && playback.FreezeLastFrame > 0 {
		v.add(path, "loop and freeze_last_frame are mutually exclusive", nil)
	}
	if !playback.Loop && rateValid && rangesValid && playback.FreezeLastFrame <= timelineDuration {
		playDuration := timelineDuration - playback.FreezeLastFrame
		expected := float64(sourceDuration) / playback.Rate
		if math.Abs(expected-float64(playDuration)) > float64(time.Microsecond) {
			v.add(path, "source duration divided by rate, plus freeze_last_frame, must equal timeline duration", nil)
		}
	}
}

func (v *projectValidator) validateClipAudio(path string, audio ClipAudio) {
	if !finite(audio.Gain) || audio.Gain < 0 {
		v.add(path+".gain", "must be finite and non-negative", nil)
	}
	if !audio.Enabled && audio.Gain != 0 {
		v.add(path+".gain", "must be zero when audio is disabled", nil)
	}
}

func (v *projectValidator) validateAudioTrack(path string, track AudioTrack, projectDuration Duration) {
	v.addObjectID(path+".id", track.ID)
	if track.Kind != AudioTrackKindBGM && track.Kind != AudioTrackKindVoice {
		v.add(path+".kind", fmt.Sprintf("has unsupported value %q", track.Kind), nil)
	}
	v.validateLocalSource(path+".source", track.Source)
	v.validateRange(path+".source_range", track.SourceRange, false)
	timelineEnd, timelineValid := v.validateRange(path+".timeline_range", track.TimelineRange, false)
	if timelineValid && projectDuration > 0 && timelineEnd > projectDuration {
		v.add(path+".timeline_range", "must not extend beyond the video sequence", nil)
	}
	if !track.Loop && track.TimelineRange.Duration > track.SourceRange.Duration {
		v.add(path+".timeline_range.duration", "must not exceed source duration when loop is false", nil)
	}
	if !finite(track.Gain) || track.Gain < 0 {
		v.add(path+".gain", "must be finite and non-negative", nil)
	}
	if v.validateDuration(path+".fade_in", track.FadeIn, true) && track.FadeIn > track.TimelineRange.Duration {
		v.add(path+".fade_in", "must not exceed timeline duration", nil)
	}
	if v.validateDuration(path+".fade_out", track.FadeOut, true) && track.FadeOut > track.TimelineRange.Duration {
		v.add(path+".fade_out", "must not exceed timeline duration", nil)
	}
}

func (v *projectValidator) validateLayer(path string, layer Layer, projectDuration Duration) {
	v.addObjectID(path+".id", layer.ID)
	layerEnd, layerRangeValid := v.validateRange(path+".range", layer.Range, false)
	if layerRangeValid && projectDuration > 0 && layerEnd > projectDuration {
		v.add(path+".range", "must not extend beyond the video sequence", nil)
	}

	payloads := 0
	if layer.Image != nil {
		payloads++
	}
	if layer.Media != nil {
		payloads++
	}
	if layer.Subtitle != nil {
		payloads++
	}
	if payloads != 1 {
		v.add(path, fmt.Sprintf("must contain exactly one payload, got %d", payloads), nil)
	}

	switch layer.Kind {
	case LayerKindImage:
		if layer.Image == nil {
			v.add(path+".image", "is required when kind is image", nil)
		}
	case LayerKindMedia:
		if layer.Media == nil {
			v.add(path+".media", "is required when kind is media", nil)
		}
	case LayerKindSubtitle:
		if layer.Subtitle == nil {
			v.add(path+".subtitle", "is required when kind is subtitle", nil)
		}
	default:
		v.add(path+".kind", fmt.Sprintf("has unsupported value %q", layer.Kind), nil)
	}

	if layer.Image != nil {
		v.validateImageLayer(path+".image", *layer.Image)
	}
	if layer.Media != nil {
		v.validateMediaLayer(path+".media", *layer.Media)
	}
	if layer.Subtitle != nil {
		v.validateSubtitleLayer(path+".subtitle", *layer.Subtitle, layer.Range, layerRangeValid)
	}
}

func (v *projectValidator) validateMediaLayer(path string, layer MediaLayer) {
	switch layer.Kind {
	case MediaKindImage:
		if layer.Loop {
			v.add(path+".loop", "must be false for a static image", nil)
		}
	case MediaKindAnimation, MediaKindVideo:
		if !layer.Loop {
			v.add(path+".loop", "must be true for animation and video layers", nil)
		}
	default:
		v.add(path+".kind", fmt.Sprintf("has unsupported value %q", layer.Kind), nil)
	}
	v.validateLocalSource(path+".source", layer.Source)
	v.validateGeometry(path+".geometry", layer.Geometry)
	if !finite(layer.Opacity) || layer.Opacity < 0 || layer.Opacity > 1 {
		v.add(path+".opacity", "must be between 0 and 1", nil)
	}
	if !finite(layer.RotationDegrees) {
		v.add(path+".rotation_degrees", "must be finite", nil)
	}
}

func (v *projectValidator) validateImageLayer(path string, layer ImageLayer) {
	v.validateLocalSource(path+".source", layer.Source)
	v.validateGeometry(path+".geometry", layer.Geometry)
	if !finite(layer.Opacity) || layer.Opacity < 0 || layer.Opacity > 1 {
		v.add(path+".opacity", "must be between 0 and 1", nil)
	}
	if !finite(layer.RotationDegrees) {
		v.add(path+".rotation_degrees", "must be finite", nil)
	}
}

func (v *projectValidator) validateSubtitleLayer(path string, layer SubtitleLayer, parentRange TimeRange, parentRangeValid bool) {
	v.validateGeometry(path+".region", layer.Region)
	v.validateSubtitleStyle(path+".style", layer.Style)
	if layer.Opacity != nil && (!finite(*layer.Opacity) || *layer.Opacity < 0 || *layer.Opacity > 1) {
		v.add(path+".opacity", "must be between 0 and 1", nil)
	}
	if !finite(layer.RotationDegrees) {
		v.add(path+".rotation_degrees", "must be finite", nil)
	}
	if len(layer.Cues) == 0 {
		v.add(path+".cues", "must contain at least one cue", nil)
	}
	parentEnd, _ := parentRange.End()
	for index, cue := range layer.Cues {
		cuePath := fmt.Sprintf("%s.cues[%d]", path, index)
		v.addObjectID(cuePath+".id", cue.ID)
		cueEnd, cueRangeValid := v.validateRange(cuePath+".range", cue.Range, false)
		if parentRangeValid && cueRangeValid && (cue.Range.Start < parentRange.Start || cueEnd > parentEnd) {
			v.add(cuePath+".range", "must be contained by the subtitle layer range", nil)
		}
		if strings.TrimSpace(cue.Text) == "" {
			v.add(cuePath+".text", "must not be empty", nil)
		}
	}
}

func (v *projectValidator) validateSubtitleStyle(path string, style SubtitleStyle) {
	if strings.TrimSpace(style.FontFamily) == "" && style.Font == nil {
		v.add(path+".font_family", "or font must be provided", nil)
	}
	if style.Font != nil {
		v.validateLocalSource(path+".font", *style.Font)
	}
	v.validateLength(path+".font_size", style.FontSize, false, true)
	if strings.TrimSpace(style.Color) == "" {
		v.add(path+".color", "must not be empty", nil)
	}
	if style.StrokeWidth.Unit != "" || style.StrokeWidth.Value != 0 {
		v.validateLength(path+".stroke_width", style.StrokeWidth, false, false)
		if style.StrokeWidth.Value > 0 && strings.TrimSpace(style.StrokeColor) == "" {
			v.add(path+".stroke_color", "must not be empty when stroke width is positive", nil)
		}
	}
	switch style.Align {
	case TextAlignLeft, TextAlignCenter, TextAlignRight:
	default:
		v.add(path+".align", fmt.Sprintf("has unsupported value %q", style.Align), nil)
	}
}

func (v *projectValidator) validateGeometry(path string, geometry Geometry) {
	switch geometry.Anchor {
	case AnchorTopLeft, AnchorTopRight, AnchorBottomLeft, AnchorBottomRight, AnchorCenter:
	default:
		v.add(path+".anchor", fmt.Sprintf("has unsupported value %q", geometry.Anchor), nil)
	}
	v.validateLength(path+".x", geometry.X, true, false)
	v.validateLength(path+".y", geometry.Y, true, false)
	v.validateLength(path+".width", geometry.Width, false, true)
	v.validateLength(path+".height", geometry.Height, false, true)
}

func (v *projectValidator) validateLength(path string, length Length, allowNegative, requirePositive bool) {
	if !finite(length.Value) {
		v.add(path+".value", "must be finite", nil)
		return
	}
	switch length.Unit {
	case LengthUnitPixel:
		if !allowNegative && length.Value < 0 {
			v.add(path+".value", "must be non-negative", nil)
		}
	case LengthUnitPercent:
		if length.Value < 0 || length.Value > 100 {
			v.add(path+".value", "percentage must be between 0 and 100", nil)
		}
	default:
		v.add(path+".unit", fmt.Sprintf("has unsupported value %q", length.Unit), nil)
	}
	if requirePositive && length.Value <= 0 {
		v.add(path+".value", "must be positive", nil)
	}
}

func (v *projectValidator) validateLocalSource(path string, source LocalSource) {
	if strings.TrimSpace(string(source.ID)) == "" {
		v.add(path+".id", "must not be empty", nil)
	}
	if strings.TrimSpace(source.Path) == "" {
		v.add(path+".path", "must not be empty", nil)
	} else if !filepath.IsAbs(source.Path) {
		v.add(path+".path", "must be an absolute local path", nil)
	}
	if source.Fingerprint.Size <= 0 {
		v.add(path+".fingerprint.size", "must be positive", nil)
	}
	if source.Fingerprint.ModifiedUnixNano == 0 {
		v.add(path+".fingerprint.modified_unix_nano", "must not be zero", nil)
	}
	if source.Fingerprint.SHA256 != "" {
		decoded, err := hex.DecodeString(source.Fingerprint.SHA256)
		if err != nil || len(decoded) != 32 {
			v.add(path+".fingerprint.sha256", "must be a 64-character hexadecimal SHA-256", nil)
		}
	}
}

func (v *projectValidator) validateFitMode(path string, fit FitMode) {
	switch fit {
	case FitModeCover, FitModeContain, FitModeStretch:
	default:
		v.add(path, fmt.Sprintf("has unsupported value %q", fit), nil)
	}
}

func (v *projectValidator) validateMetadata(path string, metadata Metadata) {
	if strings.TrimSpace(metadata.TemplateFingerprint) == "" {
		v.add(path+".template_fingerprint", "must not be empty", nil)
	}
	if strings.TrimSpace(metadata.CombinationFingerprint) == "" {
		v.add(path+".combination_fingerprint", "must not be empty", nil)
	}
	if len(metadata.Selections) == 0 {
		v.add(path+".selections", "must contain at least one selection", nil)
	}
	dimensions := make(map[ID]string)
	for index, selection := range metadata.Selections {
		selectionPath := fmt.Sprintf("%s.selections[%d]", path, index)
		switch selection.Kind {
		case SelectionKindVideo, SelectionKindTransition, SelectionKindBGM, SelectionKindVoice:
		default:
			v.add(selectionPath+".kind", fmt.Sprintf("has unsupported value %q", selection.Kind), nil)
		}
		if strings.TrimSpace(string(selection.DimensionID)) == "" {
			v.add(selectionPath+".dimension_id", "must not be empty", nil)
		} else if previous, exists := dimensions[selection.DimensionID]; exists {
			v.add(selectionPath+".dimension_id", fmt.Sprintf("duplicates dimension declared at %s", previous), nil)
		} else {
			dimensions[selection.DimensionID] = selectionPath + ".dimension_id"
		}
		if strings.TrimSpace(string(selection.OptionID)) == "" {
			v.add(selectionPath+".option_id", "must not be empty", nil)
		}
		if (selection.Kind == SelectionKindVideo || selection.Kind == SelectionKindBGM || selection.Kind == SelectionKindVoice) && strings.TrimSpace(selection.AssetFingerprint) == "" {
			v.add(selectionPath+".asset_fingerprint", "must not be empty for media selections", nil)
		}
	}

	constraints := make(map[string]string)
	for index, constraint := range metadata.Constraints {
		constraintPath := fmt.Sprintf("%s.constraints[%d]", path, index)
		if strings.TrimSpace(constraint.ID) == "" {
			v.add(constraintPath+".id", "must not be empty", nil)
		} else if previous, exists := constraints[constraint.ID]; exists {
			v.add(constraintPath+".id", fmt.Sprintf("duplicates constraint declared at %s", previous), nil)
		} else {
			constraints[constraint.ID] = constraintPath + ".id"
		}
		if strings.TrimSpace(constraint.Fingerprint) == "" {
			v.add(constraintPath+".fingerprint", "must not be empty", nil)
		}
	}
}

func (v *projectValidator) validateRange(path string, value TimeRange, allowEmpty bool) (Duration, bool) {
	valid := true
	if !v.validateDuration(path+".start", value.Start, true) {
		valid = false
	}
	if !v.validateDuration(path+".duration", value.Duration, allowEmpty) {
		valid = false
	}
	end, err := value.End()
	if err != nil {
		v.add(path, "has invalid end", err)
		return 0, false
	}
	return end, valid
}

func (v *projectValidator) validateDuration(path string, value Duration, allowZero bool) bool {
	if value < 0 {
		v.add(path, "must be non-negative", ErrInvalidDuration)
		return false
	}
	if !allowZero && value == 0 {
		v.add(path, "must be positive", nil)
		return false
	}
	if _, err := value.Microseconds(); err != nil {
		v.add(path, "must use microsecond precision", err)
		return false
	}
	return true
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
