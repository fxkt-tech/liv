package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
	"github.com/fxkt-tech/liv/ffvmix"
)

type pathList []string

func (paths *pathList) String() string {
	return strings.Join(*paths, ",")
}

func (paths *pathList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("path must not be empty")
	}
	*paths = append(*paths, value)
	return nil
}

// mixExample owns the complete example lifecycle. Its methods move the same
// object from CLI configuration to Template, Generator, and persisted output.
type mixExample struct {
	baseDir   string
	outputDir string
	count     int
	seed      uint64
	opening   pathList
	body      pathList
	bgms      pathList
	watermark string
	subtitle  string

	template  *ffvmix.Template
	generator *ffvmix.Generator
}

func main() {
	example := new(mixExample)
	example.parseFlags()
	if err := example.run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func (example *mixExample) parseFlags() {
	flag.StringVar(&example.baseDir, "base", ".", "base directory used to resolve relative asset paths")
	flag.StringVar(&example.outputDir, "out", "ffvmix-output", "directory for the template and generated FFcut projects")
	flag.IntVar(&example.count, "count", 3, "take at most N generated projects; this does not constrain FFVMix")
	flag.Uint64Var(&example.seed, "seed", 42, "generation seed; deterministic for the same persisted template")
	flag.Var(&example.opening, "opening", "opening-slot video path; repeat to add multiple candidates")
	flag.Var(&example.body, "body", "natural-duration body-slot video path; repeat to add multiple candidates")
	flag.Var(&example.bgms, "bgm", "background-music path; repeat to add multiple candidates (required)")
	flag.StringVar(&example.watermark, "watermark", "", "optional global watermark image path")
	flag.StringVar(&example.subtitle, "subtitle", "", "optional global structured subtitle text")
	flag.Parse()
}

func (example *mixExample) validate() error {
	switch {
	case len(example.opening) == 0:
		return fmt.Errorf("at least one -opening video is required")
	case len(example.body) == 0:
		return fmt.Errorf("at least one -body video is required")
	case len(example.bgms) == 0:
		return fmt.Errorf("at least one -bgm is required")
	case example.count <= 0:
		return fmt.Errorf("-count must be positive")
	default:
		return nil
	}
}

func (example *mixExample) run(ctx context.Context) error {
	if err := example.validate(); err != nil {
		return err
	}

	var err error
	example.baseDir, err = filepath.Abs(example.baseDir)
	if err != nil {
		return fmt.Errorf("resolve base directory: %w", err)
	}
	example.outputDir, err = filepath.Abs(example.outputDir)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(example.outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := example.buildTemplate(); err != nil {
		return err
	}
	templateJSON, err := ffvmix.MarshalTemplate(example.template)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}
	if err := example.writeJSON("template.ffvmix.json", templateJSON); err != nil {
		return err
	}

	// Compilation is FFVMix's only I/O phase. Generation below reads only the
	// immutable compiled state owned by Generator.
	compiled, err := ffvmix.Compile(ctx, example.template,
		ffvmix.WithBaseDir(example.baseDir),
		ffvmix.WithProbeConcurrency(4),
	)
	if err != nil {
		return fmt.Errorf("compile template: %w", err)
	}
	example.generator, err = ffvmix.NewGenerator(compiled,
		ffvmix.WithSeed(example.seed),
		ffvmix.WithConstraintFunc("no-draft-assets", "no-draft-assets/v1", rejectDraftAssets),
	)
	if err != nil {
		return fmt.Errorf("create generator: %w", err)
	}

	return example.generate(ctx)
}

func (example *mixExample) buildTemplate() error {
	openingDuration, err := ffcut.NewDuration(5 * time.Second)
	if err != nil {
		return err
	}
	fadeDuration, err := ffcut.NewDuration(500 * time.Millisecond)
	if err != nil {
		return err
	}

	defaults := ffvmix.DefaultSlotDefaults()
	defaults.Overflow = ffvmix.OverflowTrim
	defaults.Underflow = ffvmix.UnderflowLoop
	defaults.Trim = ffvmix.TrimCenter

	example.template = ffvmix.NewTemplate(ffvmix.TemplateConfig{
		Canvas: ffvmix.CanvasSpec{
			Width:     1080,
			Height:    1920,
			FrameRate: ffcut.FrameRate{Numerator: 30, Denominator: 1},
		},
		Background: ffvmix.BackgroundSpec{
			Kind:  ffvmix.BackgroundKindColor,
			Color: &ffvmix.ColorBackgroundSpec{Color: "#000000"},
		},
		Defaults: &defaults,
	})

	// A slot with TargetDuration is exactly that long. Longer videos are
	// center-trimmed and shorter videos are looped by the configured policies.
	opening, err := example.template.AddSlot(ffvmix.SlotConfig{
		Name:           "opening",
		TargetDuration: &openingDuration,
	})
	if err != nil {
		return fmt.Errorf("add opening slot: %w", err)
	}
	for _, path := range example.opening {
		if _, err := opening.AddVideo(ffvmix.VideoSourceConfig{Path: path}); err != nil {
			return fmt.Errorf("add opening video %q: %w", path, err)
		}
	}

	// A slot without TargetDuration keeps the selected video's natural length.
	body, err := example.template.AddSlot(ffvmix.SlotConfig{Name: "body"})
	if err != nil {
		return fmt.Errorf("add body slot: %w", err)
	}
	for _, path := range example.body {
		if _, err := body.AddVideo(ffvmix.VideoSourceConfig{Path: path}); err != nil {
			return fmt.Errorf("add body video %q: %w", path, err)
		}
	}

	join, err := example.template.AddJoin(ffvmix.JoinConfig{
		FromSlotID: opening.ID,
		ToSlotID:   body.ID,
	})
	if err != nil {
		return fmt.Errorf("add join: %w", err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind: ffcut.TransitionKindCut,
	}); err != nil {
		return fmt.Errorf("add cut transition: %w", err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind:           ffcut.TransitionKindFade,
		Duration:       fadeDuration,
		AudioCrossfade: true,
	}); err != nil {
		return fmt.Errorf("add fade transition: %w", err)
	}

	bgmGain := 0.25
	for _, path := range example.bgms {
		if _, err := example.template.AddBGM(ffvmix.BGMConfig{
			Path: path,
			Loop: true,
			Gain: &bgmGain,
		}); err != nil {
			return fmt.Errorf("add BGM %q: %w", path, err)
		}
	}

	if example.watermark != "" {
		opacity := 0.9
		if _, err := example.template.AddImageLayer(ffvmix.ImageLayerConfig{
			Timing:  ffvmix.FullOutputLayerTiming(),
			Path:    example.watermark,
			Opacity: &opacity,
			Geometry: ffcut.Geometry{
				Anchor: ffcut.AnchorTopLeft,
				X:      ffcut.Length{Value: 85, Unit: ffcut.LengthUnitPercent},
				Y:      ffcut.Length{Value: 3, Unit: ffcut.LengthUnitPercent},
				Width:  ffcut.Length{Value: 12, Unit: ffcut.LengthUnitPercent},
				Height: ffcut.Length{Value: 12, Unit: ffcut.LengthUnitPercent},
			},
		}); err != nil {
			return fmt.Errorf("add watermark: %w", err)
		}
	}

	if strings.TrimSpace(example.subtitle) != "" {
		cueDuration, err := ffcut.NewDuration(2 * time.Second)
		if err != nil {
			return err
		}
		if _, err := example.template.AddSubtitleLayer(ffvmix.SubtitleLayerConfig{
			Timing: ffvmix.FullOutputLayerTiming(),
			Region: ffcut.Geometry{
				Anchor: ffcut.AnchorTopLeft,
				X:      ffcut.Length{Value: 10, Unit: ffcut.LengthUnitPercent},
				Y:      ffcut.Length{Value: 75, Unit: ffcut.LengthUnitPercent},
				Width:  ffcut.Length{Value: 80, Unit: ffcut.LengthUnitPercent},
				Height: ffcut.Length{Value: 20, Unit: ffcut.LengthUnitPercent},
			},
			Style: ffvmix.SubtitleStyleSpec{
				FontFamily:      "sans-serif",
				FontSize:        ffcut.Length{Value: 42, Unit: ffcut.LengthUnitPixel},
				Color:           "#ffffff",
				BackgroundColor: "#00000080",
				Align:           ffcut.TextAlignCenter,
			},
			Input: ffvmix.StructuredSubtitles([]ffvmix.SubtitleCueSpec{{
				Range: ffcut.TimeRange{Start: 0, Duration: cueDuration},
				Text:  example.subtitle,
			}}),
		}); err != nil {
			return fmt.Errorf("add subtitle: %w", err)
		}
	}

	// These constraints belong to the Template. They restrict accepted history,
	// but they do not impose a total output count.
	example.template.AddMaxSimilarity(0.8)
	example.template.AddMaxVideoAssetUses(2)
	return nil
}

func (example *mixExample) generate(ctx context.Context) error {
	generated := 0
	for generated < example.count {
		result, err := example.generator.Next(ctx)
		if err != nil {
			return fmt.Errorf("generate project: %w", err)
		}
		switch result.Status {
		case ffvmix.Yielded:
			data, err := ffcut.Marshal(result.Project)
			if err != nil {
				return fmt.Errorf("marshal FFcut project: %w", err)
			}
			generated++
			name := fmt.Sprintf("project-%03d.ffcut.json", generated)
			if err := example.writeJSON(name, data); err != nil {
				return err
			}
			fmt.Printf("wrote %s\n", filepath.Join(example.outputDir, name))
		case ffvmix.BudgetExceeded:
			// Next preserves progress. Continue until a project is yielded or the
			// finite combination space is exhausted.
			continue
		case ffvmix.Exhausted:
			fmt.Printf("combination space exhausted after %d output(s)\n", generated)
			return nil
		default:
			return fmt.Errorf("unknown generation status %q", result.Status)
		}
	}

	stats := example.generator.Stats()
	fmt.Printf("took %d output(s): attempts=%d rejected=%d seed=%d\n",
		generated, stats.Attempts, stats.Rejected, example.generator.Seed())
	return nil
}

func (example *mixExample) writeJSON(name string, data []byte) error {
	path := filepath.Join(example.outputDir, name)
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, data, "", "  "); err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}
	formatted.WriteByte('\n')
	if err := os.WriteFile(path, formatted.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// rejectDraftAssets is intentionally a function: ConstraintFunc is FFVMix's
// plugin seam. The generator supplies immutable candidate and history views.
func rejectDraftAssets(candidate ffvmix.CandidateView, _ ffvmix.HistoryView) (ffvmix.Decision, error) {
	for _, video := range candidate.Videos() {
		if strings.Contains(strings.ToLower(filepath.Base(video.Path)), "draft") {
			return ffvmix.Reject("draft_asset"), nil
		}
	}
	return ffvmix.Accept(), nil
}
