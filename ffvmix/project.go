package ffvmix

import (
	"fmt"

	"github.com/fxkt-tech/liv/ffcut"
)

func (g *Generator) buildProject(selection selectedCombination, candidate CandidateView) (*ffcut.Project, error) {
	videoViews := candidate.Videos()
	if len(videoViews) != len(selection.videos) {
		return nil, fmt.Errorf("candidate video count %d does not match selection %d", len(videoViews), len(selection.videos))
	}

	clips := make([]ffcut.VideoClip, len(selection.videos))
	clipEnds := make([]ffcut.Duration, len(selection.videos))
	for index, video := range selection.videos {
		start := ffcut.Duration(0)
		if index > 0 {
			var err error
			start, err = subtractProtocolDuration(clipEnds[index-1], selection.transitions[index-1].Duration)
			if err != nil {
				return nil, fmt.Errorf("schedule slot %s: %w", g.slots[index].ID, err)
			}
		}
		timelineRange := ffcut.TimeRange{Start: start, Duration: video.Plan.TimelineDuration}
		end, err := timelineRange.End()
		if err != nil {
			return nil, fmt.Errorf("schedule slot %s: %w", g.slots[index].ID, err)
		}
		clipEnds[index] = end
		gain := video.AudioGain
		if !video.Metadata.HasAudio {
			gain = 0
		}
		clips[index] = ffcut.VideoClip{
			ID:            video.ID,
			Source:        projectSource("video", video.ID, video.Asset),
			SourceRange:   videoViews[index].SourceRange,
			TimelineRange: timelineRange,
			Playback: ffcut.Playback{
				Rate:            video.Plan.Rate,
				Loop:            video.Plan.Loop,
				FreezeLastFrame: video.Plan.FreezeLastFrame,
			},
			Fit: video.Fit,
			Audio: ffcut.ClipAudio{
				Enabled: video.Metadata.HasAudio,
				Gain:    gain,
			},
		}
	}

	transitions := make([]ffcut.Transition, len(selection.transitions))
	for index, transition := range selection.transitions {
		transitions[index] = ffcut.Transition{
			ID:         transition.ID,
			Kind:       transition.Kind,
			FromClipID: clips[index].ID,
			ToClipID:   clips[index+1].ID,
			Range: ffcut.TimeRange{
				Start:    clips[index+1].TimelineRange.Start,
				Duration: transition.Duration,
			},
			AudioCrossfade: transition.AudioCrossfade,
		}
	}

	outputDuration := clipEnds[len(clipEnds)-1]
	canvas, err := g.projectCanvas()
	if err != nil {
		return nil, err
	}
	audio, err := g.projectAudio(selection.bgm, outputDuration)
	if err != nil {
		return nil, err
	}
	layers, err := g.projectLayers(clips, outputDuration)
	if err != nil {
		return nil, err
	}

	projectID := uniqueProjectID(selection.fingerprint, clips, transitions, audio, layers)
	return &ffcut.Project{
		Version: ffcut.ProjectVersion,
		ID:      projectID,
		Canvas:  canvas,
		Video: ffcut.Sequence{
			Clips:       clips,
			Transitions: transitions,
		},
		Audio:    audio,
		Layers:   layers,
		Metadata: g.projectMetadata(selection, candidate),
	}, nil
}

func (g *Generator) projectCanvas() (ffcut.Canvas, error) {
	canvas := ffcut.Canvas{
		Width:     g.canvas.Width,
		Height:    g.canvas.Height,
		FrameRate: g.canvas.FrameRate,
		Background: ffcut.Background{
			Kind: g.background.Kind,
		},
	}
	switch g.background.Kind {
	case BackgroundKindColor:
		if g.background.Color == nil {
			return ffcut.Canvas{}, fmt.Errorf("compiled color background has no payload")
		}
		canvas.Background.Color = &ffcut.ColorBackground{Color: g.background.Color.Color}
	case BackgroundKindImage:
		if g.background.Image == nil {
			return ffcut.Canvas{}, fmt.Errorf("compiled image background has no payload")
		}
		canvas.Background.Image = &ffcut.ImageBackground{
			Source: projectSource("background", g.templateID, g.background.Image.Asset),
			Fit:    g.background.Image.Fit,
		}
	case BackgroundKindBlur:
		if g.background.Blur == nil {
			return ffcut.Canvas{}, fmt.Errorf("compiled blur background has no payload")
		}
		canvas.Background.Blur = &ffcut.BlurBackground{Sigma: g.background.Blur.Sigma}
	default:
		return ffcut.Canvas{}, fmt.Errorf("unsupported compiled background kind %q", g.background.Kind)
	}
	return canvas, nil
}

func (g *Generator) projectAudio(bgm *CompiledBGM, outputDuration ffcut.Duration) ([]ffcut.AudioTrack, error) {
	if bgm == nil {
		return nil, nil
	}
	duration, err := subtractProtocolDuration(outputDuration, bgm.TimelineStart)
	if err != nil {
		return nil, fmt.Errorf("BGM %s starts outside output: %w", bgm.ID, err)
	}
	if duration <= 0 {
		return nil, fmt.Errorf("BGM %s starts outside output", bgm.ID)
	}
	if !bgm.Loop && duration > bgm.SourceRange.Duration {
		duration = bgm.SourceRange.Duration
	}
	return []ffcut.AudioTrack{{
		ID:            bgm.ID,
		Kind:          ffcut.AudioTrackKindBGM,
		Source:        projectSource("bgm", bgm.ID, bgm.Asset),
		SourceRange:   bgm.SourceRange,
		TimelineRange: ffcut.TimeRange{Start: bgm.TimelineStart, Duration: duration},
		Loop:          bgm.Loop,
		Gain:          bgm.Gain,
		FadeIn:        bgm.FadeIn,
		FadeOut:       bgm.FadeOut,
	}}, nil
}

func (g *Generator) projectLayers(clips []ffcut.VideoClip, outputDuration ffcut.Duration) ([]ffcut.Layer, error) {
	clipBySlot := make(map[ID]ffcut.VideoClip, len(g.slots))
	for index, slot := range g.slots {
		clipBySlot[slot.ID] = clips[index]
	}
	layers := make([]ffcut.Layer, len(g.layers))
	for index, layer := range g.layers {
		resolvedRange, err := resolveLayerRange(layer.Timing, clipBySlot, outputDuration)
		if err != nil {
			return nil, fmt.Errorf("resolve layer %s: %w", layer.ID, err)
		}
		projectLayer := ffcut.Layer{ID: layer.ID, Kind: layer.Kind, Range: resolvedRange}
		switch layer.Kind {
		case LayerKindImage:
			if layer.Image == nil {
				return nil, fmt.Errorf("compiled image layer %s has no payload", layer.ID)
			}
			projectLayer.Image = &ffcut.ImageLayer{
				Source:          projectSource("layer", layer.ID, layer.Image.Asset),
				Geometry:        layer.Image.Geometry,
				Opacity:         layer.Image.Opacity,
				RotationDegrees: layer.Image.RotationDegrees,
			}
		case LayerKindSubtitle:
			if layer.Subtitle == nil {
				return nil, fmt.Errorf("compiled subtitle layer %s has no payload", layer.ID)
			}
			projectSubtitle, err := resolveSubtitleLayer(*layer.Subtitle, resolvedRange.Start)
			if err != nil {
				return nil, fmt.Errorf("resolve subtitle layer %s: %w", layer.ID, err)
			}
			projectLayer.Subtitle = &projectSubtitle
		default:
			return nil, fmt.Errorf("unsupported compiled layer kind %q", layer.Kind)
		}
		layers[index] = projectLayer
	}
	return layers, nil
}

func resolveLayerRange(timing LayerTiming, clips map[ID]ffcut.VideoClip, outputDuration ffcut.Duration) (ffcut.TimeRange, error) {
	switch timing.Kind {
	case LayerTimeFullOutput:
		return ffcut.TimeRange{Start: 0, Duration: outputDuration}, nil
	case LayerTimeAbsolute:
		if timing.Absolute == nil {
			return ffcut.TimeRange{}, fmt.Errorf("absolute timing payload is missing")
		}
		return timing.Absolute.Range, nil
	case LayerTimeSlot:
		if timing.Slot == nil {
			return ffcut.TimeRange{}, fmt.Errorf("slot timing payload is missing")
		}
		clip, exists := clips[timing.Slot.SlotID]
		if !exists {
			return ffcut.TimeRange{}, fmt.Errorf("slot %s is not selected", timing.Slot.SlotID)
		}
		start, err := addProtocolDuration(clip.TimelineRange.Start, timing.Slot.Offset)
		if err != nil {
			return ffcut.TimeRange{}, err
		}
		duration := clip.TimelineRange.Duration - timing.Slot.Offset
		if timing.Slot.Duration != nil {
			duration = *timing.Slot.Duration
		}
		return ffcut.TimeRange{Start: start, Duration: duration}, nil
	default:
		return ffcut.TimeRange{}, fmt.Errorf("unsupported timing kind %q", timing.Kind)
	}
}

func resolveSubtitleLayer(layer CompiledSubtitleLayer, absoluteStart ffcut.Duration) (ffcut.SubtitleLayer, error) {
	style := ffcut.SubtitleStyle{
		FontFamily:      layer.Style.FontFamily,
		FontSize:        layer.Style.FontSize,
		Color:           layer.Style.Color,
		BackgroundColor: layer.Style.BackgroundColor,
		Align:           layer.Style.Align,
	}
	if layer.Style.Font != nil {
		font := projectSource("font", "font", *layer.Style.Font)
		style.Font = &font
	}
	cues := make([]ffcut.SubtitleCue, len(layer.Cues))
	for index, cue := range layer.Cues {
		start, err := addProtocolDuration(absoluteStart, cue.Range.Start)
		if err != nil {
			return ffcut.SubtitleLayer{}, err
		}
		cues[index] = ffcut.SubtitleCue{
			ID:    cue.ID,
			Range: ffcut.TimeRange{Start: start, Duration: cue.Range.Duration},
			Text:  cue.Text,
		}
	}
	return ffcut.SubtitleLayer{Region: layer.Region, Style: style, Cues: cues}, nil
}

func (g *Generator) projectMetadata(selection selectedCombination, candidate CandidateView) ffcut.Metadata {
	selections := make([]ffcut.Selection, 0, len(selection.videos)+len(selection.transitions)+1)
	for index, video := range selection.videos {
		selections = append(selections, ffcut.Selection{
			Kind:             ffcut.SelectionKindVideo,
			DimensionID:      g.slots[index].ID,
			OptionID:         video.ID,
			AssetFingerprint: video.Asset.FingerprintString(),
		})
	}
	for index, transition := range selection.transitions {
		selections = append(selections, ffcut.Selection{
			Kind:        ffcut.SelectionKindTransition,
			DimensionID: g.joins[index].ID,
			OptionID:    transition.ID,
		})
	}
	if selection.bgm != nil {
		selections = append(selections, ffcut.Selection{
			Kind:             ffcut.SelectionKindBGM,
			DimensionID:      g.bgmDimensionID,
			OptionID:         selection.bgm.ID,
			AssetFingerprint: selection.bgm.Asset.FingerprintString(),
		})
	}
	constraintRecords := make([]ffcut.ConstraintRecord, len(g.constraints))
	for index, constraint := range g.constraints {
		constraintRecords[index] = ffcut.ConstraintRecord{
			ID:          constraint.ID(),
			Fingerprint: constraint.Fingerprint(),
		}
	}
	return ffcut.Metadata{
		TemplateFingerprint:    g.templateFingerprint,
		Seed:                   g.seed,
		CombinationFingerprint: candidate.Fingerprint(),
		Selections:             selections,
		Constraints:            constraintRecords,
	}
}

func projectSource(kind string, id ID, asset CompiledAsset) ffcut.LocalSource {
	return ffcut.LocalSource{
		ID:          ID("source:" + kind + ":" + string(id)),
		Path:        asset.Path,
		Fingerprint: asset.Fingerprint,
	}
}

func uniqueProjectID(
	fingerprint string,
	clips []ffcut.VideoClip,
	transitions []ffcut.Transition,
	audio []ffcut.AudioTrack,
	layers []ffcut.Layer,
) ID {
	used := make(map[ID]struct{}, len(clips)+len(transitions)+len(audio)+len(layers))
	for _, clip := range clips {
		used[clip.ID] = struct{}{}
	}
	for _, transition := range transitions {
		used[transition.ID] = struct{}{}
	}
	for _, track := range audio {
		used[track.ID] = struct{}{}
	}
	for _, layer := range layers {
		used[layer.ID] = struct{}{}
		if layer.Subtitle != nil {
			for _, cue := range layer.Subtitle.Cues {
				used[cue.ID] = struct{}{}
			}
		}
	}
	candidate := ID("project-" + fingerprint[:24])
	for {
		if _, exists := used[candidate]; !exists {
			return candidate
		}
		candidate += ":project"
	}
}

func subtractProtocolDuration(left, right ffcut.Duration) (ffcut.Duration, error) {
	if right < 0 || right > left {
		return 0, fmt.Errorf("cannot subtract %s from %s", right.Std(), left.Std())
	}
	return left - right, nil
}
