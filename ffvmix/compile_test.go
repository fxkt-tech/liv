package ffvmix

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

type fakeMediaProber struct {
	mu       sync.Mutex
	metadata map[string]probeMetadata
	errors   map[string]error
	calls    map[string]int
}

func (p *fakeMediaProber) Probe(_ context.Context, path string) (probeMetadata, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.calls == nil {
		p.calls = make(map[string]int)
	}
	name := filepath.Base(path)
	p.calls[name]++
	if err := p.errors[name]; err != nil {
		return probeMetadata{}, err
	}
	return p.metadata[name], nil
}

func (p *fakeMediaProber) callCount(name string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls[name]
}

func validProbeMetadata() map[string]probeMetadata {
	return map[string]probeMetadata{
		"a.mp4": {
			Video: &mediaStreamMetadata{Width: 1920, Height: 1080, Duration: 10 * time.Second},
			Audio: &mediaStreamMetadata{Duration: 10 * time.Second},
		},
		"b.mp4": {
			Video: &mediaStreamMetadata{Width: 1080, Height: 1920, Duration: 7 * time.Second},
		},
		"bgm.wav": {
			Audio: &mediaStreamMetadata{Duration: 30 * time.Second},
		},
		"watermark.png": {
			Video: &mediaStreamMetadata{Width: 100, Height: 50},
		},
		"background.png": {
			Video: &mediaStreamMetadata{Width: 1920, Height: 1080},
		},
	}
}

func writeTemplateAssets(t *testing.T, directory string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		absolute := filepath.Join(directory, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
			t.Fatalf("os.MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(absolute, []byte("fixture:"+path), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}
	}
}

func TestCompileProbesUniqueMediaAndBuildsImmutableResult(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	template := validTemplate(t)
	if _, err := template.Slots[1].AddVideo(VideoSourceConfig{Path: "media/a.mp4"}); err != nil {
		t.Fatalf("AddVideo() error = %v", err)
	}
	if _, err := template.Joins[0].AddTransition(TransitionConfig{
		Kind:           ffcut.TransitionKindFade,
		Duration:       templateDuration(t, 8*time.Second),
		AudioCrossfade: true,
	}); err != nil {
		t.Fatalf("AddTransition() error = %v", err)
	}
	before, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("json.Marshal(before) error = %v", err)
	}
	prober := &fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}
	otherDirectory := t.TempDir()
	t.Chdir(otherDirectory)

	compiled, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		WithProbeConcurrency(2),
		withMediaProber(prober),
	)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	after, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("json.Marshal(after) error = %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("Compile() mutated template\nbefore: %s\n after: %s", before, after)
	}
	if prober.callCount("a.mp4") != 1 || prober.callCount("b.mp4") != 1 || prober.callCount("bgm.wav") != 1 {
		t.Fatalf("probe counts = %#v, want one call per unique media", prober.calls)
	}
	if prober.callCount("watermark.png") != 1 {
		t.Fatal("image asset was not probed exactly once")
	}

	slots := compiled.Slots()
	if slots[0].Videos[0].Plan.Kind != AdaptationTrim || slots[0].Videos[0].Plan.TimelineDuration.Std() != 5*time.Second {
		t.Fatalf("first plan = %#v, want 5s trim", slots[0].Videos[0].Plan)
	}
	if slots[1].Videos[0].Metadata.HasAudio {
		t.Fatal("video without audio stream was not marked silent")
	}
	if !filepath.IsAbs(slots[0].Videos[0].Asset.Path) || !filepath.IsAbs(compiled.BGMs()[0].Asset.Path) {
		t.Fatal("compiled paths are not absolute")
	}
	if compiled.Fingerprint() == "" {
		t.Fatal("template fingerprint is empty")
	}
	if compiled.ID() != template.ID || compiled.Canvas() != template.Canvas || compiled.Defaults() != template.Defaults {
		t.Fatal("compiled template lost top-level values")
	}
	if len(slots[0].Videos[0].Asset.FingerprintString()) != 64 {
		t.Fatal("compiled asset fingerprint is not SHA-256 shaped")
	}
	joins := compiled.Joins()
	if !joins[0].IsCompatible(joins[0].Transitions[0].ID, slots[0].Videos[0].ID, slots[1].Videos[0].ID) {
		t.Fatal("valid fade was not marked compatible")
	}
	if joins[0].IsCompatible(joins[0].Transitions[1].ID, slots[0].Videos[0].ID, slots[1].Videos[0].ID) {
		t.Fatal("oversized fade was marked compatible")
	}

	originalPath := slots[0].Videos[0].Asset.Path
	originalCueText := compiled.Layers()[1].Subtitle.Cues[0].Text
	template.Slots[0].Videos[0].Path = "mutated.mp4"
	template.Layers[1].Subtitle.Input.Structured.Cues[0].Text = "mutated"
	template.Slots = nil
	slots[0].Videos[0].Asset.Path = "caller-mutation.mp4"
	returnedLayers := compiled.Layers()
	returnedLayers[1].Subtitle.Cues[0].Text = "caller mutation"
	if got := compiled.Slots()[0].Videos[0].Asset.Path; got != originalPath {
		t.Fatalf("compiled result mutated: got %q, want %q", got, originalPath)
	}
	if got := compiled.Layers()[1].Subtitle.Cues[0].Text; got != originalCueText {
		t.Fatalf("compiled cue mutated: got %q, want %q", got, originalCueText)
	}
	if prober.callCount("a.mp4") != 1 || prober.callCount("b.mp4") != 1 || prober.callCount("bgm.wav") != 1 {
		t.Fatal("querying compiled result performed additional probes")
	}
}

func TestCompileRequiresExplicitBaseDirectoryForRelativePaths(t *testing.T) {
	_, err := Compile(context.Background(), validTemplate(t),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %v, want *CompileError", err)
	}
	found := false
	for _, issue := range compileErr.Issues {
		if issue.Code == IssuePathResolution && issue.Path == "slots[0].videos[0].path" {
			found = true
		}
	}
	if !found {
		t.Fatalf("CompileError = %v, want relative-path issue", compileErr)
	}
}

func TestCompilePropagatesContextCancellation(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Compile(ctx, validTemplate(t),
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Compile() error = %v, want context.Canceled", err)
	}
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %T, want *CompileError", err)
	}
	found := false
	for _, issue := range compileErr.Issues {
		if issue.Code == IssueCanceled {
			found = true
		}
	}
	if !found {
		t.Fatalf("CompileError = %v, want canceled issue", compileErr)
	}
}

func TestCompileRejectsSlotLayerTimingOutsideAnyFeasibleVideo(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	template := validTemplate(t)
	duration := templateDuration(t, time.Second)
	template.Layers[0].Timing = SlotLayerTiming(template.Slots[0].ID, templateDuration(t, 4500*time.Millisecond), &duration)
	_, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %v, want *CompileError", err)
	}
	if !strings.Contains(compileErr.Error(), "layers[0].timing.slot") {
		t.Fatalf("CompileError = %v, want slot timing path", compileErr)
	}
}

func TestCompileRejectsAbsoluteLayerBeyondShortestOutput(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	template := validTemplate(t)
	template.Layers[0].Timing = AbsoluteLayerRange(ffcut.TimeRange{
		Start:    templateDuration(t, 10*time.Second),
		Duration: templateDuration(t, 2*time.Second),
	})
	_, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %v, want *CompileError", err)
	}
	if !strings.Contains(compileErr.Error(), "layers[0].timing.absolute.range") {
		t.Fatalf("CompileError = %v, want absolute timing path", compileErr)
	}
}

func TestCompileRejectsBGMStartingAfterShortestOutput(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	template := validTemplate(t)
	template.BGMs[0].TimelineStart = templateDuration(t, 12*time.Second)
	_, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %v, want *CompileError", err)
	}
	if !strings.Contains(compileErr.Error(), "bgms[0].timeline_start") {
		t.Fatalf("CompileError = %v, want BGM timing path", compileErr)
	}
}

func TestCompileStrictFingerprint(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	compiled, err := Compile(context.Background(), validTemplate(t),
		WithBaseDir(baseDirectory),
		WithStrictFingerprint(true),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	sha := compiled.Slots()[0].Videos[0].Asset.Fingerprint.SHA256
	if len(sha) != 64 {
		t.Fatalf("SHA256 length = %d, want 64", len(sha))
	}
}

func TestCompileImageBackground(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
		"images/background.png",
	)
	template := validTemplate(t)
	template.Background = BackgroundSpec{
		Kind:  BackgroundKindImage,
		Image: &ImageBackgroundSpec{Path: "images/background.png", Fit: ffcut.FitModeCover},
	}
	prober := &fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}
	compiled, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(prober),
	)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	background := compiled.Background()
	if background.Image == nil || !filepath.IsAbs(background.Image.Asset.Path) {
		t.Fatalf("compiled background = %#v", background)
	}
	if prober.callCount("background.png") != 1 {
		t.Fatalf("background probe count = %d, want 1", prober.callCount("background.png"))
	}
}

func TestCompileAggregatesIndependentFailures(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/b.mp4",
		"media/c.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	template := validTemplate(t)
	template.Joins[0].ToSlotID = "missing-slot"
	tooLong := ffcut.TimeRange{Duration: templateDuration(t, 3*time.Second)}
	if _, err := template.Slots[1].AddVideo(VideoSourceConfig{Path: "media/c.mp4", SourceRange: &tooLong}); err != nil {
		t.Fatalf("AddVideo() error = %v", err)
	}
	metadata := validProbeMetadata()
	metadata["b.mp4"] = probeMetadata{Audio: &mediaStreamMetadata{Duration: time.Second}}
	metadata["c.mp4"] = probeMetadata{Video: &mediaStreamMetadata{Width: 10, Height: 10, Duration: 2 * time.Second}}

	_, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: metadata, errors: make(map[string]error)}),
	)
	var compileErr *CompileError
	if !errors.As(err, &compileErr) {
		t.Fatalf("Compile() error = %v, want *CompileError", err)
	}
	wantCodes := map[IssueCode]bool{
		IssueInvalidReference: false,
		IssueFileStat:         false,
		IssueMissingVideo:     false,
		IssueSourceRange:      false,
	}
	for _, issue := range compileErr.Issues {
		if _, exists := wantCodes[issue.Code]; exists {
			wantCodes[issue.Code] = true
		}
	}
	for code, found := range wantCodes {
		if !found {
			t.Fatalf("CompileError missing code %q: %v", code, compileErr)
		}
	}
}

func TestCompileParsesSRTDuringCompilation(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	subtitlePath := filepath.Join(baseDirectory, "subtitles", "captions.srt")
	if err := os.MkdirAll(filepath.Dir(subtitlePath), 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(subtitlePath, []byte("1\n00:00:00,500 --> 00:00:01,500\nhello\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	template := validTemplate(t)
	template.Layers[1].Subtitle.Input = SRTSubtitles("subtitles/captions.srt")

	compiled, err := Compile(context.Background(), template,
		WithBaseDir(baseDirectory),
		withMediaProber(&fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}),
	)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	cues := compiled.Layers()[1].Subtitle.Cues
	if len(cues) != 1 || cues[0].Text != "hello" || cues[0].Range.Start.Std() != 500*time.Millisecond {
		t.Fatalf("compiled cues = %#v", cues)
	}
}

func TestCompileWithRealFFprobe(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe is not installed")
	}
	videoPath := filepath.Join(t.TempDir(), "video.mp4")
	command := exec.Command(
		"ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=black:s=16x16:r=10:d=0.2",
		"-c:v", "mpeg4", "-an", videoPath,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate test video: %v: %s", err, output)
	}
	template := NewTemplate(TemplateConfig{Canvas: CanvasSpec{Width: 16, Height: 16}})
	slot, err := template.AddSlot(SlotConfig{Name: "only"})
	if err != nil {
		t.Fatalf("AddSlot() error = %v", err)
	}
	if _, err := slot.AddVideo(VideoSourceConfig{Path: videoPath}); err != nil {
		t.Fatalf("AddVideo() error = %v", err)
	}

	compiled, err := Compile(context.Background(), template)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	video := compiled.Slots()[0].Videos[0]
	if video.Metadata.Duration <= 0 || video.Metadata.Width != 16 || video.Metadata.Height != 16 {
		t.Fatalf("real probe metadata = %#v", video.Metadata)
	}
	if video.Metadata.HasAudio {
		t.Fatal("video-only fixture unexpectedly has audio")
	}
}

func TestProtocolMediaDurationFloorsToMicroseconds(t *testing.T) {
	got, err := protocolMediaDuration(time.Microsecond + 999*time.Nanosecond)
	if err != nil {
		t.Fatalf("protocolMediaDuration() error = %v", err)
	}
	if got.Std() != time.Microsecond {
		t.Fatalf("protocolMediaDuration() = %s, want 1us", got.Std())
	}
}
