package ffvmix

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func Compile(ctx context.Context, template *Template, optionFunctions ...CompileOption) (*CompiledTemplate, error) {
	options := defaultCompileOptions()
	for _, option := range optionFunctions {
		if option != nil {
			option(&options)
		}
	}

	issues := make([]Issue, 0)
	if ctx == nil {
		issues = append(issues, Issue{Code: IssueCanceled, Path: "context", Message: "must not be nil", Cause: ErrInvalidTemplate})
		ctx = context.Background()
	}
	if template == nil {
		issues = append(issues, Issue{Code: IssueInvalidValue, Path: "template", Message: "is required", Cause: ErrInvalidTemplate})
		return nil, &CompileError{Issues: issues}
	}
	if err := template.Validate(); err != nil {
		var compileErr *CompileError
		if errors.As(err, &compileErr) {
			issues = append(issues, compileErr.Issues...)
		} else {
			issues = append(issues, Issue{Code: IssueInvalidValue, Path: "template", Message: "validation failed", Cause: err})
		}
	}

	resolved := collectTemplateAssets(template, options, &issues)
	inspections := inspectAssets(ctx, resolved, options, &issues)
	compiled := buildCompiledTemplate(template, resolved, inspections, &issues)
	if err := ctx.Err(); err != nil && !hasIssueCause(issues, err) {
		issues = append(issues, Issue{Code: IssueCanceled, Path: "compile", Message: "context canceled", Cause: err})
	}
	if len(issues) > 0 {
		return nil, &CompileError{Issues: issues}
	}

	fingerprint, err := templateFingerprint(template)
	if err != nil {
		return nil, &CompileError{Issues: []Issue{{Code: IssueInvalidValue, Path: "template", Message: "cannot calculate template fingerprint", Cause: err}}}
	}
	compiled.fingerprint = fingerprint
	return compiled, nil
}

func buildCompiledTemplate(
	template *Template,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	issues *[]Issue,
) *CompiledTemplate {
	compiled := &CompiledTemplate{
		id:          template.ID,
		canvas:      template.Canvas,
		defaults:    template.Defaults,
		constraints: cloneConstraints(template.Constraints),
	}
	compiled.background = compileBackground(template.Background, resolved, inspections, issues)
	compiled.slots = compileSlots(template, resolved, inspections, issues)
	compiled.bgms = compileBGMs(template, resolved, inspections, issues)
	compiled.joins = compileJoins(template, compiled.slots, issues)
	minimumOutputDuration, hasOutput := minimumCompiledOutputDuration(compiled.slots, compiled.joins)
	validateCompiledBGMs(compiled.bgms, minimumOutputDuration, hasOutput, issues)
	compiled.layers = compileLayers(template, resolved, inspections, compiled.slots, minimumOutputDuration, hasOutput, issues)
	return compiled
}

func compileBackground(
	spec BackgroundSpec,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	issues *[]Issue,
) CompiledBackground {
	background := CompiledBackground{Kind: spec.Kind}
	if spec.Color != nil {
		color := *spec.Color
		background.Color = &color
	}
	if spec.Blur != nil {
		blur := *spec.Blur
		background.Blur = &blur
	}
	if spec.Image != nil {
		path := resolved.backgroundImage
		inspection, exists := inspections[path]
		if exists && inspectionUsable(inspection) && validateCompiledImage("background.image.path", path, inspection.metadata, issues) {
			background.Image = &CompiledImageBackground{Asset: inspection.asset, Fit: spec.Image.Fit}
		}
	}
	return background
}

func compileSlots(
	template *Template,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	issues *[]Issue,
) []CompiledSlot {
	slots := make([]CompiledSlot, 0, len(template.Slots))
	for slotIndex, slot := range template.Slots {
		if slot == nil {
			continue
		}
		policy := effectiveSlotPolicy(template.Defaults, slot.Overrides)
		compiledSlot := CompiledSlot{
			ID:     slot.ID,
			Name:   slot.Name,
			Policy: policy,
			Videos: make([]CompiledVideo, 0, len(slot.Videos)),
		}
		if slot.TargetDuration != nil {
			compiledSlot.HasTargetDuration = true
			compiledSlot.TargetDuration = *slot.TargetDuration
		}
		feasible := 0
		for videoIndex, video := range slot.Videos {
			if video == nil {
				continue
			}
			path := resolved.videos[video]
			inspection, exists := inspections[path]
			if !exists || !inspectionUsable(inspection) {
				continue
			}
			fieldPath := fmt.Sprintf("slots[%d].videos[%d]", slotIndex, videoIndex)
			metadata, ok := compileVideoMetadata(fieldPath, path, inspection.metadata, issues)
			if !ok {
				continue
			}
			available, ok := compileSourceRange(fieldPath+".source_range", video.SourceRange, metadata.Duration, path, issues)
			if !ok {
				continue
			}
			fit := policy.Fit
			if video.Fit != nil {
				fit = *video.Fit
			}
			audioGain := policy.AudioGain
			if video.AudioGain != nil {
				audioGain = *video.AudioGain
			}
			plan := planAdaptation(available, slot.TargetDuration, policy)
			if plan.Feasible {
				feasible++
			}
			compiledSlot.Videos = append(compiledSlot.Videos, CompiledVideo{
				ID:        video.ID,
				Asset:     inspection.asset,
				Metadata:  metadata,
				Weight:    video.Weight,
				Fit:       fit,
				AudioGain: audioGain,
				Plan:      plan,
			})
		}
		if feasible == 0 {
			*issues = append(*issues, Issue{
				Code:    IssueNoFeasibleSource,
				Path:    fmt.Sprintf("slots[%d].videos", slotIndex),
				Message: "slot has no feasible video source",
			})
		}
		slots = append(slots, compiledSlot)
	}
	return slots
}

func compileVideoMetadata(path, localPath string, metadata probeMetadata, issues *[]Issue) (VideoMetadata, bool) {
	if metadata.Video == nil {
		*issues = append(*issues, Issue{Code: IssueMissingVideo, Path: path + ".path", LocalPath: localPath, Message: "first video stream is required"})
		return VideoMetadata{}, false
	}
	if metadata.Video.Width <= 0 || metadata.Video.Height <= 0 {
		*issues = append(*issues, Issue{Code: IssueMissingVideo, Path: path + ".path", LocalPath: localPath, Message: "first video stream must have positive dimensions"})
		return VideoMetadata{}, false
	}
	duration := metadata.Video.Duration
	if duration <= 0 {
		duration = metadata.FormatDuration
	}
	protocolDuration, err := protocolMediaDuration(duration)
	if err != nil || protocolDuration <= 0 {
		*issues = append(*issues, Issue{Code: IssueMissingVideo, Path: path + ".path", LocalPath: localPath, Message: "first video stream must have a positive duration", Cause: err})
		return VideoMetadata{}, false
	}
	result := VideoMetadata{
		Width:    metadata.Video.Width,
		Height:   metadata.Video.Height,
		Duration: protocolDuration,
	}
	if metadata.Audio != nil {
		audioDuration := metadata.Audio.Duration
		if audioDuration <= 0 {
			audioDuration = metadata.FormatDuration
		}
		if audioDuration > 0 {
			converted, err := protocolMediaDuration(audioDuration)
			if err == nil {
				result.HasAudio = true
				result.AudioDuration = converted
			}
		}
	}
	return result, true
}

func compileSourceRange(path string, configured *ffcut.TimeRange, mediaDuration ffcut.Duration, localPath string, issues *[]Issue) (ffcut.TimeRange, bool) {
	if configured == nil {
		return ffcut.TimeRange{Start: 0, Duration: mediaDuration}, true
	}
	end, err := configured.End()
	if err != nil || configured.Start < 0 || configured.Duration <= 0 || end > mediaDuration {
		*issues = append(*issues, Issue{
			Code:      IssueSourceRange,
			Path:      path,
			LocalPath: localPath,
			Message:   fmt.Sprintf("must be contained by media duration %s", mediaDuration.Std()),
			Cause:     err,
		})
		return ffcut.TimeRange{}, false
	}
	return *configured, true
}

func compileBGMs(
	template *Template,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	issues *[]Issue,
) []CompiledBGM {
	bgms := make([]CompiledBGM, 0, len(template.BGMs))
	for index, bgm := range template.BGMs {
		if bgm == nil {
			continue
		}
		path := resolved.bgms[bgm]
		inspection, exists := inspections[path]
		if !exists || !inspectionUsable(inspection) {
			continue
		}
		fieldPath := fmt.Sprintf("bgms[%d]", index)
		if inspection.metadata.Audio == nil {
			*issues = append(*issues, Issue{Code: IssueMissingAudio, Path: fieldPath + ".path", LocalPath: path, Message: "first audio stream is required"})
			continue
		}
		duration := inspection.metadata.Audio.Duration
		if duration <= 0 {
			duration = inspection.metadata.FormatDuration
		}
		protocolDuration, err := protocolMediaDuration(duration)
		if err != nil || protocolDuration <= 0 {
			*issues = append(*issues, Issue{Code: IssueMissingAudio, Path: fieldPath + ".path", LocalPath: path, Message: "first audio stream must have a positive duration", Cause: err})
			continue
		}
		sourceRange, ok := compileSourceRange(fieldPath+".source_range", bgm.SourceRange, protocolDuration, path, issues)
		if !ok {
			continue
		}
		if !bgm.Loop && (bgm.FadeIn > sourceRange.Duration || bgm.FadeOut > sourceRange.Duration) {
			*issues = append(*issues, Issue{Code: IssueInvalidValue, Path: fieldPath, LocalPath: path, Message: "fade duration must not exceed non-looping source range", Cause: ErrInvalidTemplate})
			continue
		}
		bgms = append(bgms, CompiledBGM{
			ID:            bgm.ID,
			Asset:         inspection.asset,
			SourceRange:   sourceRange,
			TimelineStart: bgm.TimelineStart,
			Loop:          bgm.Loop,
			Gain:          bgm.Gain,
			FadeIn:        bgm.FadeIn,
			FadeOut:       bgm.FadeOut,
			Weight:        bgm.Weight,
			templateIndex: index,
		})
	}
	return bgms
}

func compileLayers(
	template *Template,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	slots []CompiledSlot,
	minimumOutputDuration ffcut.Duration,
	hasOutput bool,
	issues *[]Issue,
) []CompiledLayer {
	layers := make([]CompiledLayer, 0, len(template.Layers))
	slotsByID := make(map[ID]CompiledSlot, len(slots))
	for _, slot := range slots {
		slotsByID[slot.ID] = slot
	}
	for index, layer := range template.Layers {
		if layer == nil {
			continue
		}
		fieldPath := fmt.Sprintf("layers[%d]", index)
		compiled := CompiledLayer{ID: layer.ID, Kind: layer.Kind, Timing: cloneLayerTiming(layer.Timing)}
		if layer.Image != nil {
			path := resolved.images[layer]
			inspection, exists := inspections[path]
			if exists && inspectionUsable(inspection) && validateCompiledImage(fieldPath+".image.path", path, inspection.metadata, issues) {
				compiled.Image = &CompiledImageLayer{
					Asset:           inspection.asset,
					Geometry:        layer.Image.Geometry,
					Opacity:         layer.Image.Opacity,
					RotationDegrees: layer.Image.RotationDegrees,
				}
			}
		}
		if layer.Subtitle != nil {
			compiled.Subtitle = compileSubtitleLayer(fieldPath, layer, resolved, inspections, issues)
		}
		validateCompiledLayerTiming(fieldPath, layer, compiled, slotsByID, minimumOutputDuration, hasOutput, issues)
		layers = append(layers, compiled)
	}
	return layers
}

func validateCompiledBGMs(bgms []CompiledBGM, minimumOutputDuration ffcut.Duration, hasOutput bool, issues *[]Issue) {
	if !hasOutput {
		return
	}
	for _, bgm := range bgms {
		path := fmt.Sprintf("bgms[%d]", bgm.templateIndex)
		if bgm.TimelineStart >= minimumOutputDuration {
			*issues = append(*issues, Issue{
				Code:    IssueInvalidValue,
				Path:    path + ".timeline_start",
				Message: fmt.Sprintf("must be before minimum output duration %s", minimumOutputDuration.Std()),
				Cause:   ErrInvalidTemplate,
			})
			continue
		}
		playDuration := minimumOutputDuration - bgm.TimelineStart
		if !bgm.Loop && bgm.SourceRange.Duration < playDuration {
			playDuration = bgm.SourceRange.Duration
		}
		if bgm.FadeIn > playDuration || bgm.FadeOut > playDuration {
			*issues = append(*issues, Issue{
				Code:    IssueInvalidValue,
				Path:    path,
				Message: fmt.Sprintf("fade duration must fit minimum BGM play duration %s", playDuration.Std()),
				Cause:   ErrInvalidTemplate,
			})
		}
	}
}

func validateCompiledImage(fieldPath, localPath string, metadata probeMetadata, issues *[]Issue) bool {
	if metadata.Video == nil {
		*issues = append(*issues, Issue{Code: IssueMissingVideo, Path: fieldPath, LocalPath: localPath, Message: "image must have a visual stream"})
		return false
	}
	if metadata.Video.Width <= 0 || metadata.Video.Height <= 0 {
		*issues = append(*issues, Issue{Code: IssueMissingVideo, Path: fieldPath, LocalPath: localPath, Message: "image must have positive dimensions"})
		return false
	}
	return true
}

func compileSubtitleLayer(
	fieldPath string,
	layer *LayerSpec,
	resolved *resolvedTemplateAssets,
	inspections map[string]assetInspection,
	issues *[]Issue,
) *CompiledSubtitleLayer {
	spec := layer.Subtitle
	compiled := &CompiledSubtitleLayer{
		Region: spec.Region,
		Style: CompiledSubtitleStyle{
			FontFamily:      spec.Style.FontFamily,
			FontSize:        spec.Style.FontSize,
			Color:           spec.Style.Color,
			BackgroundColor: spec.Style.BackgroundColor,
			Align:           spec.Style.Align,
		},
	}
	if fontPath := resolved.fonts[layer]; fontPath != "" {
		if inspection, exists := inspections[fontPath]; exists && inspectionUsable(inspection) {
			font := inspection.asset
			compiled.Style.Font = &font
		}
	}

	switch spec.Input.Kind {
	case SubtitleInputStructured:
		if spec.Input.Structured != nil {
			compiled.Cues = make([]NormalizedCue, len(spec.Input.Structured.Cues))
			for index, cue := range spec.Input.Structured.Cues {
				compiled.Cues[index] = NormalizedCue(cue)
			}
		}
	case SubtitleInputSRT, SubtitleInputASS:
		path := resolved.subtitles[layer]
		inspection, exists := inspections[path]
		if !exists || !inspectionUsable(inspection) {
			break
		}
		cues, err := parseSubtitleFile(spec.Input.Kind, path, layer.ID)
		if err != nil {
			*issues = append(*issues, Issue{Code: IssueSubtitleParse, Path: fieldPath + ".subtitle.input", LocalPath: path, Message: "cannot parse subtitle file", Cause: err})
			break
		}
		compiled.Cues = cues
	}
	return compiled
}

func validateCompiledLayerTiming(
	fieldPath string,
	spec *LayerSpec,
	compiled CompiledLayer,
	slotsByID map[ID]CompiledSlot,
	minimumOutputDuration ffcut.Duration,
	hasOutput bool,
	issues *[]Issue,
) {
	var localDuration ffcut.Duration
	hasLocalDuration := false
	if spec.Timing.Absolute != nil {
		localDuration = spec.Timing.Absolute.Range.Duration
		hasLocalDuration = true
		if end, err := spec.Timing.Absolute.Range.End(); err == nil && hasOutput && end > minimumOutputDuration {
			*issues = append(*issues, Issue{
				Code:    IssueInvalidValue,
				Path:    fieldPath + ".timing.absolute.range",
				Message: fmt.Sprintf("must fit minimum output duration %s", minimumOutputDuration.Std()),
				Cause:   ErrInvalidTemplate,
			})
		}
	}
	if spec.Timing.FullOutput != nil && hasOutput && compiled.Subtitle != nil {
		validateCueBounds(fieldPath, compiled.Subtitle.Cues, minimumOutputDuration, issues)
	}
	if spec.Timing.Slot != nil {
		timing := spec.Timing.Slot
		if timing.Duration != nil {
			localDuration = *timing.Duration
			hasLocalDuration = true
		}
		if slot, exists := slotsByID[timing.SlotID]; exists {
			var minimumRemaining ffcut.Duration
			hasRemaining := false
			timingFits := true
			for _, video := range slot.Videos {
				if !video.Plan.Feasible {
					continue
				}
				end := timing.Offset
				if timing.Duration != nil {
					resolvedEnd, err := (ffcut.TimeRange{Start: timing.Offset, Duration: *timing.Duration}).End()
					if err != nil {
						continue
					}
					end = resolvedEnd
				}
				if (timing.Duration == nil && timing.Offset >= video.Plan.TimelineDuration) || (timing.Duration != nil && end > video.Plan.TimelineDuration) {
					*issues = append(*issues, Issue{
						Code:    IssueInvalidReference,
						Path:    fieldPath + ".timing.slot",
						Message: fmt.Sprintf("does not fit feasible video %q duration %s", video.ID, video.Plan.TimelineDuration.Std()),
						Cause:   ErrInvalidTemplate,
					})
					timingFits = false
					break
				}
				if timing.Duration == nil {
					remaining := video.Plan.TimelineDuration - timing.Offset
					if !hasRemaining || remaining < minimumRemaining {
						minimumRemaining = remaining
						hasRemaining = true
					}
				}
			}
			if timingFits && timing.Duration == nil && compiled.Subtitle != nil && hasRemaining {
				validateCueBounds(fieldPath, compiled.Subtitle.Cues, minimumRemaining, issues)
			}
		}
	}
	if !hasLocalDuration || compiled.Subtitle == nil {
		return
	}
	validateCueBounds(fieldPath, compiled.Subtitle.Cues, localDuration, issues)
}

func validateCueBounds(fieldPath string, cues []NormalizedCue, maximum ffcut.Duration, issues *[]Issue) {
	for index, cue := range cues {
		end, err := cue.Range.End()
		if err != nil || end > maximum {
			*issues = append(*issues, Issue{
				Code:    IssueInvalidValue,
				Path:    fmt.Sprintf("%s.subtitle.cues[%d].range", fieldPath, index),
				Message: fmt.Sprintf("must fit resolved layer duration %s", maximum.Std()),
				Cause:   err,
			})
		}
	}
}

func minimumCompiledOutputDuration(slots []CompiledSlot, joins []CompiledJoin) (ffcut.Duration, bool) {
	if len(slots) == 0 {
		return 0, false
	}
	current := make(map[ID]ffcut.Duration)
	for _, video := range slots[0].Videos {
		if video.Plan.Feasible {
			current[video.ID] = video.Plan.TimelineDuration
		}
	}
	for slotIndex := 1; slotIndex < len(slots); slotIndex++ {
		if slotIndex-1 >= len(joins) {
			return 0, false
		}
		join := joins[slotIndex-1]
		next := make(map[ID]ffcut.Duration)
		for _, to := range slots[slotIndex].Videos {
			if !to.Plan.Feasible {
				continue
			}
			for fromID, accumulated := range current {
				for _, transition := range join.Transitions {
					if !join.IsCompatible(transition.ID, fromID, to.ID) {
						continue
					}
					end, err := (ffcut.TimeRange{Start: accumulated, Duration: to.Plan.TimelineDuration}).End()
					if err != nil || transition.Duration > end {
						continue
					}
					candidate := end - transition.Duration
					if previous, exists := next[to.ID]; !exists || candidate < previous {
						next[to.ID] = candidate
					}
				}
			}
		}
		current = next
	}
	var minimum ffcut.Duration
	found := false
	for _, duration := range current {
		if !found || duration < minimum {
			minimum = duration
			found = true
		}
	}
	return minimum, found
}

func compileJoins(template *Template, slots []CompiledSlot, issues *[]Issue) []CompiledJoin {
	joins := make([]CompiledJoin, 0, len(template.Joins))
	for index, join := range template.Joins {
		if join == nil {
			continue
		}
		compiled := CompiledJoin{
			ID:            join.ID,
			FromSlotID:    join.FromSlotID,
			ToSlotID:      join.ToSlotID,
			Transitions:   make([]CompiledTransition, 0, len(join.Transitions)),
			compatibility: make(map[compatibilityKey]bool),
		}
		for _, transition := range join.Transitions {
			if transition == nil {
				continue
			}
			compiled.Transitions = append(compiled.Transitions, CompiledTransition{
				ID:             transition.ID,
				Kind:           transition.Kind,
				Duration:       transition.Duration,
				AudioCrossfade: transition.AudioCrossfade,
				Weight:         transition.Weight,
			})
		}
		compatibleCount := 0
		if index < len(slots)-1 {
			for _, transition := range compiled.Transitions {
				for _, from := range slots[index].Videos {
					for _, to := range slots[index+1].Videos {
						compatible := from.Plan.Feasible && to.Plan.Feasible &&
							transition.Duration <= from.Plan.TimelineDuration &&
							transition.Duration <= to.Plan.TimelineDuration
						key := compatibilityKey{TransitionID: transition.ID, FromVideoID: from.ID, ToVideoID: to.ID}
						compiled.compatibility[key] = compatible
						if compatible {
							compatibleCount++
						}
					}
				}
			}
		}
		if compatibleCount == 0 {
			*issues = append(*issues, Issue{
				Code:    IssueNoFeasibleTransition,
				Path:    fmt.Sprintf("joins[%d].transitions", index),
				Message: "join has no transition compatible with any feasible adjacent video pair",
			})
		}
		joins = append(joins, compiled)
	}
	return joins
}

func protocolMediaDuration(value time.Duration) (ffcut.Duration, error) {
	if value <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	// Floor positive probe durations so the protocol never claims source time
	// beyond the media boundary after reducing precision to microseconds.
	value = value.Truncate(time.Microsecond)
	if value <= 0 {
		return 0, fmt.Errorf("duration rounds to zero microseconds")
	}
	return ffcut.NewDuration(value)
}

func inspectionUsable(value assetInspection) bool {
	return value.statErr == nil && value.fingerprintErr == nil && value.probeErr == nil && value.asset.Path != ""
}

func templateFingerprint(template *Template) (string, error) {
	data, err := json.Marshal(template)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func hasIssueCause(issues []Issue, target error) bool {
	for _, issue := range issues {
		if errors.Is(issue.Cause, target) {
			return true
		}
	}
	return false
}
