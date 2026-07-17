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

type options struct {
	baseDir   string
	outputDir string
	count     int
	seed      uint64
	opening   pathList
	body      pathList
	bgms      pathList
	watermark string
	subtitle  string
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	config := parseOptions()
	if err := config.validate(); err != nil {
		return err
	}

	baseDir, err := filepath.Abs(config.baseDir)
	if err != nil {
		return fmt.Errorf("resolve base directory: %w", err)
	}
	outputDir, err := filepath.Abs(config.outputDir)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	template, err := buildTemplate(config)
	if err != nil {
		return err
	}
	templateJSON, err := ffvmix.MarshalTemplate(template)
	if err != nil {
		return fmt.Errorf("marshal template: %w", err)
	}
	if err := writeJSON(filepath.Join(outputDir, "template.ffvmix.json"), templateJSON); err != nil {
		return err
	}

	// Compile is FFVMix's only I/O phase: it resolves local paths, stats and
	// fingerprints files, and probes each unique media file with ffprobe.
	compiled, err := ffvmix.Compile(ctx, template,
		ffvmix.WithBaseDir(baseDir),
		ffvmix.WithProbeConcurrency(4),
	)
	if err != nil {
		return fmt.Errorf("compile template: %w", err)
	}

	// A custom constraint is a pure plugin function. This example rejects
	// assets whose base name contains "draft"; HistoryView is also available
	// for rules based on previously accepted results.
	generator, err := ffvmix.NewGenerator(compiled,
		ffvmix.WithSeed(config.seed),
		ffvmix.WithConstraintFunc("no-draft-assets", "no-draft-assets/v1", rejectDraftAssets),
	)
	if err != nil {
		return fmt.Errorf("create generator: %w", err)
	}

	return writeProjects(ctx, generator, outputDir, config.count)
}

func parseOptions() options {
	var config options
	flag.StringVar(&config.baseDir, "base", ".", "base directory used to resolve relative asset paths")
	flag.StringVar(&config.outputDir, "out", "ffvmix-output", "directory for the template and generated FFcut projects")
	flag.IntVar(&config.count, "count", 3, "take at most N generated projects; this does not constrain FFVMix")
	flag.Uint64Var(&config.seed, "seed", 42, "generation seed; deterministic for the same persisted template")
	flag.Var(&config.opening, "opening", "opening-slot video path; repeat to add multiple candidates")
	flag.Var(&config.body, "body", "natural-duration body-slot video path; repeat to add multiple candidates")
	flag.Var(&config.bgms, "bgm", "background-music path; repeat to add multiple candidates (required)")
	flag.StringVar(&config.watermark, "watermark", "", "optional global watermark image path")
	flag.StringVar(&config.subtitle, "subtitle", "", "optional global structured subtitle text")
	flag.Parse()
	return config
}

func (config options) validate() error {
	switch {
	case len(config.opening) == 0:
		return fmt.Errorf("at least one -opening video is required")
	case len(config.body) == 0:
		return fmt.Errorf("at least one -body video is required")
	case len(config.bgms) == 0:
		return fmt.Errorf("at least one -bgm is required")
	case config.count <= 0:
		return fmt.Errorf("-count must be positive")
	default:
		return nil
	}
}

func buildTemplate(config options) (*ffvmix.Template, error) {
	openingDuration, err := ffcut.NewDuration(5 * time.Second)
	if err != nil {
		return nil, err
	}
	fadeDuration, err := ffcut.NewDuration(500 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	defaults := ffvmix.DefaultSlotDefaults()
	defaults.Overflow = ffvmix.OverflowTrim
	defaults.Underflow = ffvmix.UnderflowLoop
	defaults.Trim = ffvmix.TrimCenter

	template := ffvmix.NewTemplate(ffvmix.TemplateConfig{
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
	opening, err := template.AddSlot(ffvmix.SlotConfig{
		Name:           "opening",
		TargetDuration: &openingDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("add opening slot: %w", err)
	}
	if err := addVideos(opening, config.opening); err != nil {
		return nil, err
	}

	// A slot without TargetDuration keeps the selected video's natural length.
	body, err := template.AddSlot(ffvmix.SlotConfig{Name: "body"})
	if err != nil {
		return nil, fmt.Errorf("add body slot: %w", err)
	}
	if err := addVideos(body, config.body); err != nil {
		return nil, err
	}

	join, err := template.AddJoin(ffvmix.JoinConfig{
		FromSlotID: opening.ID,
		ToSlotID:   body.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("add join: %w", err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind: ffcut.TransitionKindCut,
	}); err != nil {
		return nil, fmt.Errorf("add cut transition: %w", err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind:           ffcut.TransitionKindFade,
		Duration:       fadeDuration,
		AudioCrossfade: true,
	}); err != nil {
		return nil, fmt.Errorf("add fade transition: %w", err)
	}

	bgmGain := 0.25
	for _, path := range config.bgms {
		if _, err := template.AddBGM(ffvmix.BGMConfig{
			Path: path,
			Loop: true,
			Gain: &bgmGain,
		}); err != nil {
			return nil, fmt.Errorf("add BGM %q: %w", path, err)
		}
	}

	if config.watermark != "" {
		opacity := 0.9
		if _, err := template.AddImageLayer(ffvmix.ImageLayerConfig{
			Timing:  ffvmix.FullOutputLayerTiming(),
			Path:    config.watermark,
			Opacity: &opacity,
			Geometry: percentGeometry(
				ffcut.AnchorTopLeft,
				85, 3, 12, 12,
			),
		}); err != nil {
			return nil, fmt.Errorf("add watermark: %w", err)
		}
	}

	if strings.TrimSpace(config.subtitle) != "" {
		cueDuration, err := ffcut.NewDuration(2 * time.Second)
		if err != nil {
			return nil, err
		}
		if _, err := template.AddSubtitleLayer(ffvmix.SubtitleLayerConfig{
			Timing: ffvmix.FullOutputLayerTiming(),
			Region: percentGeometry(ffcut.AnchorTopLeft, 10, 75, 80, 20),
			Style: ffvmix.SubtitleStyleSpec{
				FontFamily:      "sans-serif",
				FontSize:        ffcut.Length{Value: 42, Unit: ffcut.LengthUnitPixel},
				Color:           "#ffffff",
				BackgroundColor: "#00000080",
				Align:           ffcut.TextAlignCenter,
			},
			Input: ffvmix.StructuredSubtitles([]ffvmix.SubtitleCueSpec{{
				Range: ffcut.TimeRange{Start: 0, Duration: cueDuration},
				Text:  config.subtitle,
			}}),
		}); err != nil {
			return nil, fmt.Errorf("add subtitle: %w", err)
		}
	}

	// Built-in history constraints are persisted in the template. They filter
	// accepted outputs without setting a total generation count.
	template.AddMaxSimilarity(0.8)
	template.AddMaxVideoAssetUses(2)
	return template, nil
}

func addVideos(slot *ffvmix.Slot, paths []string) error {
	for _, path := range paths {
		if _, err := slot.AddVideo(ffvmix.VideoSourceConfig{Path: path}); err != nil {
			return fmt.Errorf("add video %q: %w", path, err)
		}
	}
	return nil
}

func percentGeometry(anchor ffcut.Anchor, x, y, width, height float64) ffcut.Geometry {
	return ffcut.Geometry{
		Anchor: anchor,
		X:      ffcut.Length{Value: x, Unit: ffcut.LengthUnitPercent},
		Y:      ffcut.Length{Value: y, Unit: ffcut.LengthUnitPercent},
		Width:  ffcut.Length{Value: width, Unit: ffcut.LengthUnitPercent},
		Height: ffcut.Length{Value: height, Unit: ffcut.LengthUnitPercent},
	}
}

func rejectDraftAssets(candidate ffvmix.CandidateView, _ ffvmix.HistoryView) (ffvmix.Decision, error) {
	for _, video := range candidate.Videos() {
		if strings.Contains(strings.ToLower(filepath.Base(video.Path)), "draft") {
			return ffvmix.Reject("draft_asset"), nil
		}
	}
	return ffvmix.Accept(), nil
}

func writeProjects(ctx context.Context, generator *ffvmix.Generator, outputDir string, count int) error {
	generated := 0
	for generated < count {
		result, err := generator.Next(ctx)
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
			path := filepath.Join(outputDir, fmt.Sprintf("project-%03d.ffcut.json", generated))
			if err := writeJSON(path, data); err != nil {
				return err
			}
			fmt.Printf("wrote %s\n", path)
		case ffvmix.BudgetExceeded:
			// Next preserves progress, so continue until one project is yielded or
			// the finite combination space is exhausted.
			continue
		case ffvmix.Exhausted:
			fmt.Printf("combination space exhausted after %d output(s)\n", generated)
			return nil
		default:
			return fmt.Errorf("unknown generation status %q", result.Status)
		}
	}

	stats := generator.Stats()
	fmt.Printf("took %d output(s): attempts=%d rejected=%d seed=%d\n",
		generated, stats.Attempts, stats.Rejected, generator.Seed())
	return nil
}

func writeJSON(path string, data []byte) error {
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
