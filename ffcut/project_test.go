package ffcut

import (
	"encoding/json"
	"testing"
	"time"
)

func validProject() *Project {
	return &Project{
		Version: ProjectVersion,
		ID:      "project-1",
		Canvas: Canvas{
			Width:     1920,
			Height:    1080,
			FrameRate: FrameRate{Numerator: 30, Denominator: 1},
			Background: Background{
				Kind:  BackgroundKindColor,
				Color: &ColorBackground{Color: "#000000"},
			},
		},
		Video: Sequence{
			Clips: []VideoClip{
				{
					ID:            "clip-a",
					Source:        testSource("source-a", "/media/a.mp4", 1000, 1),
					SourceRange:   testRange(0, 5*time.Second),
					TimelineRange: testRange(0, 5*time.Second),
					Playback:      Playback{Rate: 1},
					Fit:           FitModeCover,
					Audio:         ClipAudio{Enabled: true, Gain: 1},
				},
				{
					ID:            "clip-b",
					Source:        testSource("source-b", "/media/b.mp4", 2000, 2),
					SourceRange:   testRange(0, 5*time.Second),
					TimelineRange: testRange(4*time.Second, 5*time.Second),
					Playback:      Playback{Rate: 1},
					Fit:           FitModeContain,
					Audio:         ClipAudio{Enabled: false, Gain: 0},
				},
			},
			Transitions: []Transition{
				{
					ID:             "transition-a-b",
					Kind:           TransitionKindFade,
					FromClipID:     "clip-a",
					ToClipID:       "clip-b",
					Range:          testRange(4*time.Second, time.Second),
					AudioCrossfade: true,
				},
			},
		},
		Audio: []AudioTrack{
			{
				ID:            "bgm-track",
				Kind:          AudioTrackKindBGM,
				Source:        testSource("bgm-source", "/media/bgm.wav", 3000, 3),
				SourceRange:   testRange(0, 9*time.Second),
				TimelineRange: testRange(0, 9*time.Second),
				Gain:          0.8,
				FadeIn:        duration(time.Second),
				FadeOut:       duration(time.Second),
			},
		},
		Layers: []Layer{
			{
				ID:    "watermark-layer",
				Kind:  LayerKindImage,
				Range: testRange(0, 9*time.Second),
				Image: &ImageLayer{
					Source:   testSource("watermark-source", "/media/watermark.png", 4000, 4),
					Geometry: testGeometry(Length{Value: 20, Unit: LengthUnitPercent}, Length{Value: 120, Unit: LengthUnitPixel}),
					Opacity:  1,
				},
			},
			{
				ID:    "subtitle-layer",
				Kind:  LayerKindSubtitle,
				Range: testRange(0, 9*time.Second),
				Subtitle: &SubtitleLayer{
					Region: testGeometry(Length{Value: 80, Unit: LengthUnitPercent}, Length{Value: 20, Unit: LengthUnitPercent}),
					Style: SubtitleStyle{
						FontFamily: "sans-serif",
						FontSize:   Length{Value: 32, Unit: LengthUnitPixel},
						Color:      "#FFFFFF",
						Align:      TextAlignCenter,
					},
					Cues: []SubtitleCue{
						{ID: "cue-1", Range: testRange(time.Second, 2*time.Second), Text: "first"},
						{ID: "cue-2", Range: testRange(5*time.Second, 2*time.Second), Text: "second"},
					},
				},
			},
		},
		Metadata: Metadata{
			TemplateFingerprint:    "template-sha256",
			Seed:                   42,
			CombinationFingerprint: "combination-sha256",
			Selections: []Selection{
				{Kind: SelectionKindVideo, DimensionID: "slot-a", OptionID: "candidate-a", AssetFingerprint: "asset-a"},
				{Kind: SelectionKindTransition, DimensionID: "join-a-b", OptionID: "fade"},
				{Kind: SelectionKindVideo, DimensionID: "slot-b", OptionID: "candidate-b", AssetFingerprint: "asset-b"},
				{Kind: SelectionKindBGM, DimensionID: "bgm-pool", OptionID: "bgm-a", AssetFingerprint: "asset-bgm"},
			},
			Constraints: []ConstraintRecord{
				{ID: "max-similarity", Fingerprint: "constraint-sha256"},
			},
		},
	}
}

func testSource(id ID, path string, size, modified int64) LocalSource {
	return LocalSource{
		ID:   id,
		Path: path,
		Fingerprint: MediaFingerprint{
			Size:             size,
			ModifiedUnixNano: modified,
		},
	}
}

func testRange(start, length time.Duration) TimeRange {
	return TimeRange{Start: duration(start), Duration: duration(length)}
}

func duration(value time.Duration) Duration {
	parsed, err := NewDuration(value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func testGeometry(width, height Length) Geometry {
	return Geometry{
		Anchor: AnchorTopLeft,
		X:      Length{Value: 10, Unit: LengthUnitPixel},
		Y:      Length{Value: 10, Unit: LengthUnitPixel},
		Width:  width,
		Height: height,
	}
}

func cloneProject(t *testing.T, project *Project) *Project {
	t.Helper()
	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var cloned Project
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return &cloned
}
