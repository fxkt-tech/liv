package ffvmix

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func TestCompileToGeneratorIntegrationPerformsNoGenerationIO(t *testing.T) {
	baseDirectory := t.TempDir()
	writeTemplateAssets(t, baseDirectory,
		"media/a.mp4",
		"media/b.mp4",
		"audio/bgm.wav",
		"images/watermark.png",
	)
	prober := &fakeMediaProber{metadata: validProbeMetadata(), errors: make(map[string]error)}
	compiled, err := Compile(context.Background(), validTemplate(t),
		WithBaseDir(baseDirectory),
		withMediaProber(prober),
	)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	before := map[string]int{
		"a.mp4":         prober.callCount("a.mp4"),
		"b.mp4":         prober.callCount("b.mp4"),
		"bgm.wav":       prober.callCount("bgm.wav"),
		"watermark.png": prober.callCount("watermark.png"),
	}

	project := nextProject(t, newTestGenerator(t, compiled, WithSeed(88)))
	if err := project.Validate(); err != nil {
		t.Fatalf("Project.Validate() error = %v", err)
	}
	encoded, err := ffcut.Marshal(project)
	if err != nil {
		t.Fatalf("ffcut.Marshal() error = %v", err)
	}
	decoded, err := ffcut.Unmarshal(encoded)
	if err != nil {
		t.Fatalf("ffcut.Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(decoded, project) {
		t.Fatalf("generated Project changed across protocol round-trip")
	}
	for name, count := range before {
		if got := prober.callCount(name); got != count {
			t.Fatalf("probe count for %s = %d after generation, want %d", name, got, count)
		}
	}
}

func TestGeneratorExhaustiveTraversalIsUniqueAndDeterministic(t *testing.T) {
	first, firstStats := collectProjects(t, newTestGenerator(t, generatorFixture(), WithSeed(42)))
	second, _ := collectProjects(t, newTestGenerator(t, generatorFixture(), WithSeed(42)))
	if got, want := len(first), 2*2*2*2; got != want {
		t.Fatalf("yield count = %d, want %d", got, want)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different order\nfirst:  %v\nsecond: %v", first, second)
	}
	seen := make(map[string]struct{}, len(first))
	for _, fingerprint := range first {
		if _, exists := seen[fingerprint]; exists {
			t.Fatalf("duplicate combination %s", fingerprint)
		}
		seen[fingerprint] = struct{}{}
	}
	if firstStats.Attempts != 16 || firstStats.Yielded != 16 || firstStats.Rejected != 0 {
		t.Fatalf("stats = %+v", firstStats)
	}

	changed := false
	for seed := uint64(43); seed < 60; seed++ {
		other, _ := collectProjects(t, newTestGenerator(t, generatorFixture(), WithSeed(seed)))
		if !reflect.DeepEqual(first, other) {
			changed = true
		}
		sort.Strings(other)
		expected := append([]string(nil), first...)
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, other) {
			t.Fatalf("seed %d changed the unconstrained terminal set", seed)
		}
		if changed {
			break
		}
	}
	if !changed {
		t.Fatal("different seeds did not alter order")
	}
}

func TestGeneratorOptionsAndActualSeed(t *testing.T) {
	if _, err := NewGenerator(nil); !errors.Is(err, ErrInvalidGenerator) {
		t.Fatalf("NewGenerator(nil) error = %v", err)
	}
	if _, err := NewGenerator(generatorFixture(), WithSearchBudget(0)); !errors.Is(err, ErrInvalidGenerator) {
		t.Fatalf("zero budget error = %v", err)
	}
	if _, err := NewGenerator(generatorFixture(), WithConstraintFunc("", "fingerprint", func(CandidateView, HistoryView) (Decision, error) {
		return Accept(), nil
	})); !errors.Is(err, ErrInvalidGenerator) {
		t.Fatalf("invalid function constraint error = %v", err)
	}

	compiled := generatorFixture()
	compiled.constraints = []ConstraintSpec{{
		ID: "duplicate", Kind: ConstraintMaxSimilarity,
		MaxSimilarity: &MaxSimilaritySpec{Maximum: 1},
	}}
	if _, err := NewGenerator(compiled, WithConstraintFunc("duplicate", "custom", func(CandidateView, HistoryView) (Decision, error) {
		return Accept(), nil
	})); !errors.Is(err, ErrInvalidGenerator) {
		t.Fatalf("duplicate constraint error = %v", err)
	}

	generator := newTestGenerator(t, generatorFixture())
	project := nextProject(t, generator)
	if project.Metadata.Seed != generator.Seed() {
		t.Fatalf("metadata seed = %d, Generator.Seed() = %d", project.Metadata.Seed, generator.Seed())
	}
	stats := generator.Stats()
	stats.Rejections["caller-mutation"] = 99
	if generator.Stats().Rejections["caller-mutation"] != 0 {
		t.Fatal("Stats() returned mutable rejection state")
	}
}

func TestGeneratorBudgetExceededPreservesProgress(t *testing.T) {
	generator := newTestGenerator(t, generatorFixture(),
		WithSeed(1),
		WithSearchBudget(2),
		WithConstraintFunc("reject-all", "v1", func(CandidateView, HistoryView) (Decision, error) {
			return Reject("custom_reject"), nil
		}),
	)
	first, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("first Next() error = %v", err)
	}
	if first.Status != BudgetExceeded || first.Stats.Attempts != 2 {
		t.Fatalf("first result = %+v", first)
	}
	second, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("second Next() error = %v", err)
	}
	if second.Status != BudgetExceeded || second.Stats.Attempts != 4 {
		t.Fatalf("second result = %+v", second)
	}
}

func TestConstraintErrorConsumesAttemptButDoesNotCommitHistory(t *testing.T) {
	sentinel := errors.New("constraint failed")
	generator := newTestGenerator(t, generatorFixture(),
		WithSeed(2),
		WithConstraintFunc("error", "v1", func(CandidateView, HistoryView) (Decision, error) {
			return Decision{}, sentinel
		}),
	)
	_, err := generator.Next(context.Background())
	if !errors.Is(err, ErrConstraintCheck) || !errors.Is(err, sentinel) {
		t.Fatalf("Next() error = %v", err)
	}
	if got := len(generator.history); got != 0 {
		t.Fatalf("history length = %d, want 0", got)
	}
	stats := generator.Stats()
	if stats.Attempts != 1 || stats.Yielded != 0 || stats.Rejected != 0 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestBuiltInConstraintsUseAcceptedHistoryOnly(t *testing.T) {
	tests := []struct {
		name       string
		constraint ConstraintSpec
		reason     string
	}{
		{
			name: "similarity",
			constraint: ConstraintSpec{
				ID: "similarity", Kind: ConstraintMaxSimilarity,
				MaxSimilarity: &MaxSimilaritySpec{Maximum: 0},
			},
			reason: ReasonMaxSimilarity,
		},
		{
			name: "video uses",
			constraint: ConstraintSpec{
				ID: "video-uses", Kind: ConstraintMaxVideoAssetUses,
				MaxVideoAssetUses: &MaxVideoAssetUsesSpec{Maximum: 1},
			},
			reason: ReasonMaxVideoAssetUses,
		},
		{
			name: "BGM uses",
			constraint: ConstraintSpec{
				ID: "bgm-uses", Kind: ConstraintMaxBGMUses,
				MaxBGMUses: &MaxBGMUsesSpec{Maximum: 1},
			},
			reason: ReasonMaxBGMUses,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compiled := generatorFixture()
			compiled.constraints = []ConstraintSpec{test.constraint}
			projects, stats := collectProjects(t, newTestGenerator(t, compiled, WithSeed(9)))
			if got, want := len(projects), 2; got != want {
				t.Fatalf("yield count = %d, want %d; stats = %+v", got, want, stats)
			}
			if stats.Rejections[test.reason] == 0 {
				t.Fatalf("rejections = %+v, want reason %q", stats.Rejections, test.reason)
			}
		})
	}
}

func TestGeneratorRecordsEngineRejections(t *testing.T) {
	t.Run("infeasible video", func(t *testing.T) {
		compiled := generatorFixture()
		compiled.slots[0].Videos[1].Plan.Feasible = false
		compiled.slots[0].Videos[1].Plan.Kind = AdaptationInfeasible
		projects, stats := collectProjects(t, newTestGenerator(t, compiled, WithSeed(3)))
		if len(projects) != 8 || stats.Rejections[ReasonInfeasibleVideo] != 8 {
			t.Fatalf("projects = %d, stats = %+v", len(projects), stats)
		}
	})

	t.Run("incompatible transition", func(t *testing.T) {
		compiled := generatorFixture()
		key := compatibilityKey{
			TransitionID: compiled.joins[0].Transitions[1].ID,
			FromVideoID:  compiled.slots[0].Videos[0].ID,
			ToVideoID:    compiled.slots[1].Videos[0].ID,
		}
		compiled.joins[0].compatibility[key] = false
		projects, stats := collectProjects(t, newTestGenerator(t, compiled, WithSeed(3)))
		if len(projects) != 14 || stats.Rejections[ReasonIncompatibleTransition] != 2 {
			t.Fatalf("projects = %d, stats = %+v", len(projects), stats)
		}
	})
}

func TestGeneratedProjectResolvesTimelineBGMAndLayers(t *testing.T) {
	compiled := generatorFixture()
	compiled.slots[0].Videos = compiled.slots[0].Videos[:1]
	compiled.slots[1].Videos = compiled.slots[1].Videos[:1]
	compiled.joins[0].Transitions = compiled.joins[0].Transitions[1:]
	compiled.joins[0].compatibility = map[compatibilityKey]bool{{
		TransitionID: compiled.joins[0].Transitions[0].ID,
		FromVideoID:  compiled.slots[0].Videos[0].ID,
		ToVideoID:    compiled.slots[1].Videos[0].ID,
	}: true}
	compiled.bgms = compiled.bgms[:1]
	compiled.bgms[0].TimelineStart = generatorDuration(time.Second)
	compiled.slots[1].Videos[0].Metadata.HasAudio = false
	compiled.layers = []CompiledLayer{
		{
			ID: "watermark", Kind: LayerKindImage, Timing: FullOutputLayerTiming(),
			Image: &CompiledImageLayer{
				Asset:    generatorAsset("/watermark.png", 20),
				Geometry: generatorGeometry(),
				Opacity:  0.8,
			},
		},
		{
			ID: "subtitle", Kind: LayerKindSubtitle,
			Timing: SlotLayerTiming("slot-b", generatorDuration(time.Second), generatorDurationPointer(2*time.Second)),
			Subtitle: &CompiledSubtitleLayer{
				Region: generatorGeometry(),
				Style: CompiledSubtitleStyle{
					FontFamily: "sans-serif",
					FontSize:   ffcut.Length{Value: 24, Unit: ffcut.LengthUnitPixel},
					Color:      "#ffffff",
					Align:      ffcut.TextAlignCenter,
				},
				Cues: []NormalizedCue{{
					ID: "cue", Range: generatorRange(500*time.Millisecond, time.Second), Text: "hello",
				}},
			},
		},
	}

	generator := newTestGenerator(t, compiled, WithSeed(11))
	result, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if result.Status != Yielded {
		t.Fatalf("status = %s", result.Status)
	}
	project := result.Project
	if err := project.Validate(); err != nil {
		t.Fatalf("Project.Validate() error = %v", err)
	}
	if got, want := project.Video.Clips[1].TimelineRange.Start, generatorDuration(4*time.Second); got != want {
		t.Fatalf("second clip start = %s, want %s", got.Std(), want.Std())
	}
	if got, want := project.Video.Transitions[0].Range, generatorRange(4*time.Second, time.Second); got != want {
		t.Fatalf("transition range = %+v, want %+v", got, want)
	}
	if got, want := project.Audio[0].TimelineRange, generatorRange(time.Second, 8*time.Second); got != want {
		t.Fatalf("BGM range = %+v, want %+v", got, want)
	}
	if project.Video.Clips[1].Audio.Enabled || project.Video.Clips[1].Audio.Gain != 0 {
		t.Fatalf("silent clip audio = %+v", project.Video.Clips[1].Audio)
	}
	if got, want := project.Layers[0].Range, generatorRange(0, 9*time.Second); got != want {
		t.Fatalf("full layer range = %+v, want %+v", got, want)
	}
	if got, want := project.Layers[1].Range, generatorRange(5*time.Second, 2*time.Second); got != want {
		t.Fatalf("slot layer range = %+v, want %+v", got, want)
	}
	if got, want := project.Layers[1].Subtitle.Cues[0].Range, generatorRange(5500*time.Millisecond, time.Second); got != want {
		t.Fatalf("cue range = %+v, want %+v", got, want)
	}
	if project.Metadata.Seed != 11 || len(project.Metadata.Selections) != 4 || project.Metadata.CombinationFingerprint == "" {
		t.Fatalf("metadata = %+v", project.Metadata)
	}
}

func TestRandomTrimIsSeededAndBounded(t *testing.T) {
	compiled := generatorFixture()
	compiled.slots = compiled.slots[:1]
	compiled.joins = nil
	compiled.bgms = nil
	compiled.slots[0].Videos = compiled.slots[0].Videos[:1]
	compiled.slots[0].Videos[0].Plan = AdaptationPlan{
		Kind:              AdaptationTrim,
		Feasible:          true,
		AvailableRange:    generatorRange(0, 10*time.Second),
		SourceDuration:    generatorDuration(5 * time.Second),
		TimelineDuration:  generatorDuration(5 * time.Second),
		Rate:              1,
		TrimMode:          TrimRandom,
		MaximumTrimOffset: generatorDuration(5 * time.Second),
	}

	first := nextProject(t, newTestGenerator(t, compiled, WithSeed(100)))
	second := nextProject(t, newTestGenerator(t, compiled, WithSeed(100)))
	if first.Video.Clips[0].SourceRange != second.Video.Clips[0].SourceRange {
		t.Fatalf("same seed trim differs: %+v vs %+v", first.Video.Clips[0].SourceRange, second.Video.Clips[0].SourceRange)
	}
	start := first.Video.Clips[0].SourceRange.Start
	if start < 0 || start > generatorDuration(5*time.Second) {
		t.Fatalf("trim start = %s, outside [0,5s]", start.Std())
	}
	changed := false
	for seed := uint64(101); seed < 120; seed++ {
		other := nextProject(t, newTestGenerator(t, compiled, WithSeed(seed)))
		if other.Video.Clips[0].SourceRange.Start != start {
			changed = true
			break
		}
	}
	if !changed {
		t.Fatal("different seeds did not alter random trim offset")
	}
}

func TestGeneratorCancellationDoesNotAdvance(t *testing.T) {
	generator := newTestGenerator(t, generatorFixture(), WithSeed(4))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := generator.Next(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Next() error = %v", err)
	}
	if stats := generator.Stats(); stats.Attempts != 0 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestSyntheticBGMDimensionAvoidsPersistedIDCollision(t *testing.T) {
	compiled := generatorFixture()
	compiled.slots[0].ID = "template:bgm"
	compiled.joins[0].FromSlotID = compiled.slots[0].ID
	project := nextProject(t, newTestGenerator(t, compiled, WithSeed(13)))
	seen := make(map[ID]struct{})
	for _, selection := range project.Metadata.Selections {
		if _, exists := seen[selection.DimensionID]; exists {
			t.Fatalf("duplicate metadata dimension %q", selection.DimensionID)
		}
		seen[selection.DimensionID] = struct{}{}
	}
	if _, exists := seen["template:bgm:pool"]; !exists {
		t.Fatalf("metadata dimensions = %+v", project.Metadata.Selections)
	}
}

func TestUniqueProjectIDAvoidsProtocolObjectCollision(t *testing.T) {
	fingerprint := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	colliding := ID("project-" + fingerprint[:24])
	got := uniqueProjectID(fingerprint, []ffcut.VideoClip{{ID: colliding}}, nil, nil, nil)
	if got == colliding {
		t.Fatalf("project ID %q collides with clip ID", got)
	}
}

func TestConcurrentNextFailsClearly(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	generator := newTestGenerator(t, generatorFixture(),
		WithSeed(5),
		WithConstraintFunc("blocking", "v1", func(CandidateView, HistoryView) (Decision, error) {
			once.Do(func() { close(started) })
			<-release
			return Accept(), nil
		}),
	)
	errCh := make(chan error, 1)
	go func() {
		_, err := generator.Next(context.Background())
		errCh <- err
	}()
	<-started
	_, err := generator.Next(context.Background())
	if !errors.Is(err, ErrConcurrentNext) {
		t.Fatalf("concurrent Next() error = %v", err)
	}
	close(release)
	if err := <-errCh; err != nil {
		t.Fatalf("first Next() error = %v", err)
	}
}

func generatorFixture() *CompiledTemplate {
	slotA := CompiledSlot{ID: "slot-a", Videos: []CompiledVideo{
		generatorVideo("video-a", "/a.mp4", 1, 5*time.Second, 9),
		generatorVideo("video-b", "/b.mp4", 2, 5*time.Second, 1),
	}}
	slotB := CompiledSlot{ID: "slot-b", Videos: []CompiledVideo{
		generatorVideo("video-c", "/c.mp4", 3, 5*time.Second, 1),
		generatorVideo("video-d", "/d.mp4", 4, 5*time.Second, 1),
	}}
	cut := CompiledTransition{ID: "cut", Kind: ffcut.TransitionKindCut, Weight: 1}
	fade := CompiledTransition{
		ID: "fade", Kind: ffcut.TransitionKindFade, Duration: generatorDuration(time.Second),
		AudioCrossfade: true, Weight: 1,
	}
	join := CompiledJoin{
		ID: "join", FromSlotID: slotA.ID, ToSlotID: slotB.ID,
		Transitions:   []CompiledTransition{cut, fade},
		compatibility: make(map[compatibilityKey]bool),
	}
	for _, transition := range join.Transitions {
		for _, from := range slotA.Videos {
			for _, to := range slotB.Videos {
				join.compatibility[compatibilityKey{TransitionID: transition.ID, FromVideoID: from.ID, ToVideoID: to.ID}] = true
			}
		}
	}
	return &CompiledTemplate{
		id:          "template",
		fingerprint: "template-fingerprint",
		canvas: CanvasSpec{
			Width: 1920, Height: 1080,
			FrameRate: ffcut.FrameRate{Numerator: 30, Denominator: 1},
		},
		background: CompiledBackground{
			Kind: BackgroundKindColor, Color: &ColorBackgroundSpec{Color: "#000000"},
		},
		slots: []CompiledSlot{slotA, slotB},
		joins: []CompiledJoin{join},
		bgms: []CompiledBGM{
			generatorBGM("bgm-a", "/bgm-a.mp3", 10, 9),
			generatorBGM("bgm-b", "/bgm-b.mp3", 11, 1),
		},
	}
}

func generatorVideo(id, path string, modified int64, duration time.Duration, weight float64) CompiledVideo {
	protocolDuration := generatorDuration(duration)
	return CompiledVideo{
		ID: ID(id), Asset: generatorAsset(path, modified), Weight: weight,
		Fit: ffcut.FitModeCover, AudioGain: 1,
		Metadata: VideoMetadata{Width: 1920, Height: 1080, Duration: protocolDuration, HasAudio: true, AudioDuration: protocolDuration},
		Plan: AdaptationPlan{
			Kind: AdaptationNatural, Feasible: true,
			AvailableRange: generatorRange(0, duration), SourceDuration: protocolDuration,
			TimelineDuration: protocolDuration, Rate: 1,
		},
	}
}

func generatorBGM(id, path string, modified int64, weight float64) CompiledBGM {
	return CompiledBGM{
		ID: ID(id), Asset: generatorAsset(path, modified),
		SourceRange: generatorRange(0, 20*time.Second), Gain: 0.5, Weight: weight,
	}
}

func generatorAsset(path string, modified int64) CompiledAsset {
	return CompiledAsset{
		Path: path,
		Fingerprint: ffcut.MediaFingerprint{
			Size: 100, ModifiedUnixNano: modified,
		},
	}
}

func generatorGeometry() ffcut.Geometry {
	return ffcut.Geometry{
		Anchor: ffcut.AnchorTopLeft,
		X:      ffcut.Length{Value: 0, Unit: ffcut.LengthUnitPixel},
		Y:      ffcut.Length{Value: 0, Unit: ffcut.LengthUnitPixel},
		Width:  ffcut.Length{Value: 100, Unit: ffcut.LengthUnitPixel},
		Height: ffcut.Length{Value: 100, Unit: ffcut.LengthUnitPixel},
	}
}

func generatorRange(start, duration time.Duration) ffcut.TimeRange {
	return ffcut.TimeRange{Start: generatorDuration(start), Duration: generatorDuration(duration)}
}

func generatorDuration(value time.Duration) ffcut.Duration {
	result, err := ffcut.NewDuration(value)
	if err != nil {
		panic(err)
	}
	return result
}

func generatorDurationPointer(value time.Duration) *ffcut.Duration {
	duration := generatorDuration(value)
	return &duration
}

func newTestGenerator(t *testing.T, compiled *CompiledTemplate, options ...GeneratorOption) *Generator {
	t.Helper()
	generator, err := NewGenerator(compiled, options...)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	return generator
}

func nextProject(t *testing.T, generator *Generator) *ffcut.Project {
	t.Helper()
	result, err := generator.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if result.Status != Yielded || result.Project == nil {
		t.Fatalf("Next() result = %+v", result)
	}
	return result.Project
}

func collectProjects(t *testing.T, generator *Generator) ([]string, GenerationStats) {
	t.Helper()
	var fingerprints []string
	for {
		result, err := generator.Next(context.Background())
		if err != nil {
			t.Fatalf("Next() error = %v", err)
		}
		switch result.Status {
		case Yielded:
			if result.Project == nil {
				t.Fatal("Yielded result has nil Project")
			}
			if err := result.Project.Validate(); err != nil {
				t.Fatalf("generated Project validation error = %v", err)
			}
			fingerprints = append(fingerprints, result.Project.Metadata.CombinationFingerprint)
		case BudgetExceeded:
			continue
		case Exhausted:
			return fingerprints, result.Stats
		default:
			t.Fatalf("unknown generation status %q", result.Status)
		}
	}
}
