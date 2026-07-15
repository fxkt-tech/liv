package ffvmix

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func validTemplate(t *testing.T) *Template {
	t.Helper()
	defaults := DefaultSlotDefaults()
	defaults.Overflow = OverflowTrim
	defaults.Underflow = UnderflowLoop
	template := NewTemplate(TemplateConfig{
		Canvas:   CanvasSpec{Width: 1920, Height: 1080},
		Defaults: &defaults,
	})

	target := templateDuration(t, 5*time.Second)
	first, err := template.AddSlot(SlotConfig{Name: "opening", TargetDuration: &target})
	if err != nil {
		t.Fatalf("AddSlot() error = %v", err)
	}
	if _, err := first.AddVideo(VideoSourceConfig{Path: "media/a.mp4"}); err != nil {
		t.Fatalf("AddVideo() error = %v", err)
	}
	second, err := template.AddSlot(SlotConfig{Name: "body"})
	if err != nil {
		t.Fatalf("AddSlot() error = %v", err)
	}
	if _, err := second.AddVideo(VideoSourceConfig{Path: "media/b.mp4", Weight: 2}); err != nil {
		t.Fatalf("AddVideo() error = %v", err)
	}

	join, err := template.AddJoin(JoinConfig{FromSlotID: first.ID, ToSlotID: second.ID})
	if err != nil {
		t.Fatalf("AddJoin() error = %v", err)
	}
	if _, err := join.AddTransition(TransitionConfig{Kind: ffcut.TransitionKindFade, Duration: templateDuration(t, time.Second), AudioCrossfade: true}); err != nil {
		t.Fatalf("AddTransition() error = %v", err)
	}

	if _, err := template.AddBGM(BGMConfig{Path: "audio/bgm.wav", Loop: true}); err != nil {
		t.Fatalf("AddBGM() error = %v", err)
	}
	if _, err := template.AddImageLayer(ImageLayerConfig{
		Timing:   FullOutputLayerTiming(),
		Path:     "images/watermark.png",
		Geometry: testTemplateGeometry(),
	}); err != nil {
		t.Fatalf("AddImageLayer() error = %v", err)
	}
	if _, err := template.AddSubtitleLayer(SubtitleLayerConfig{
		Timing: FullOutputLayerTiming(),
		Region: testTemplateGeometry(),
		Style: SubtitleStyleSpec{
			FontFamily: "sans-serif",
			FontSize:   ffcut.Length{Value: 32, Unit: ffcut.LengthUnitPixel},
			Color:      "#ffffff",
			Align:      ffcut.TextAlignCenter,
		},
		Input: StructuredSubtitles([]SubtitleCueSpec{{
			Range: ffcut.TimeRange{Start: 0, Duration: templateDuration(t, 2*time.Second)},
			Text:  "hello",
		}}),
	}); err != nil {
		t.Fatalf("AddSubtitleLayer() error = %v", err)
	}
	template.AddMaxSimilarity(0.5)
	template.AddMaxVideoAssetUses(2)
	template.AddMaxBGMUses(1)
	return template
}

func templateDuration(t *testing.T, value time.Duration) ffcut.Duration {
	t.Helper()
	duration, err := ffcut.NewDuration(value)
	if err != nil {
		t.Fatalf("ffcut.NewDuration() error = %v", err)
	}
	return duration
}

func testTemplateGeometry() ffcut.Geometry {
	return ffcut.Geometry{
		Anchor: ffcut.AnchorTopLeft,
		X:      ffcut.Length{Value: 10, Unit: ffcut.LengthUnitPixel},
		Y:      ffcut.Length{Value: 10, Unit: ffcut.LengthUnitPixel},
		Width:  ffcut.Length{Value: 50, Unit: ffcut.LengthUnitPercent},
		Height: ffcut.Length{Value: 20, Unit: ffcut.LengthUnitPercent},
	}
}

func cloneTemplate(t *testing.T, template *Template) *Template {
	t.Helper()
	data, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var cloned Template
	if err := json.Unmarshal(data, &cloned); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	return &cloned
}

func TestTemplateJSONRoundTripPreservesIDsAndOrder(t *testing.T) {
	want := validTemplate(t)
	data, err := MarshalTemplate(want)
	if err != nil {
		t.Fatalf("MarshalTemplate() error = %v", err)
	}
	got, err := UnmarshalTemplate(data)
	if err != nil {
		t.Fatalf("UnmarshalTemplate() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\n got: %#v\nwant: %#v", got, want)
	}
	if got.Slots[0].Name != "opening" || got.Slots[1].Name != "body" {
		t.Fatalf("slot order changed: %q, %q", got.Slots[0].Name, got.Slots[1].Name)
	}
}

func TestTemplateConstructorsGeneratePersistentIDsAndDefaults(t *testing.T) {
	template := validTemplate(t)
	ids := []ID{
		template.ID,
		template.Slots[0].ID,
		template.Slots[0].Videos[0].ID,
		template.Joins[0].ID,
		template.Joins[0].Transitions[0].ID,
		template.BGMs[0].ID,
		template.Layers[0].ID,
		template.Layers[1].Subtitle.Input.Structured.Cues[0].ID,
	}
	seen := make(map[ID]struct{})
	for _, id := range ids {
		if id == "" {
			t.Fatal("constructor generated an empty ID")
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("constructor generated duplicate ID %q", id)
		}
		seen[id] = struct{}{}
	}
	if template.Canvas.FrameRate != (ffcut.FrameRate{Numerator: 30, Denominator: 1}) {
		t.Fatalf("frame rate = %#v, want 30/1", template.Canvas.FrameRate)
	}
	if template.Slots[0].Videos[0].Weight != 1 || template.BGMs[0].Weight != 1 {
		t.Fatal("constructor did not apply default weight")
	}
}

func TestNewTemplateOwnsBackgroundPayload(t *testing.T) {
	color := &ColorBackgroundSpec{Color: "#112233"}
	template := NewTemplate(TemplateConfig{
		Canvas:     CanvasSpec{Width: 10, Height: 10},
		Background: BackgroundSpec{Kind: BackgroundKindColor, Color: color},
	})
	color.Color = "#ffffff"
	if template.Background.Color.Color != "#112233" {
		t.Fatalf("template background = %q, want constructor-owned copy", template.Background.Color.Color)
	}
}

func TestUnmarshalTemplateRejectsMissingIDsAndUnknownFields(t *testing.T) {
	data, err := MarshalTemplate(validTemplate(t))
	if err != nil {
		t.Fatalf("MarshalTemplate() error = %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	raw["id"] = ""
	missingID, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, err := UnmarshalTemplate(missingID); !errors.Is(err, ErrInvalidTemplate) {
		t.Fatalf("UnmarshalTemplate(missing ID) error = %v, want ErrInvalidTemplate", err)
	}

	unknown := strings.TrimSuffix(string(data), "}") + `,"unknown":true}`
	if _, err := UnmarshalTemplate([]byte(unknown)); !errors.Is(err, ErrInvalidTemplate) {
		t.Fatalf("UnmarshalTemplate(unknown) error = %v, want ErrInvalidTemplate", err)
	}
}

func TestTemplateValidateAggregatesStructuralIssues(t *testing.T) {
	template := cloneTemplate(t, validTemplate(t))
	template.Canvas.Width = 0
	template.Slots[0].ID = ""
	template.Slots[0].Videos[0].Weight = -1
	template.Joins[0].ToSlotID = "missing"
	template.Layers[0].Timing.FullOutput = nil

	err := template.Validate()
	var compileErr *CompileError
	if !errors.As(err, &compileErr) || !errors.Is(err, ErrInvalidTemplate) {
		t.Fatalf("Validate() error = %v, want invalid CompileError", err)
	}
	if len(compileErr.Issues) < 5 {
		t.Fatalf("len(Issues) = %d, want at least 5", len(compileErr.Issues))
	}
}
