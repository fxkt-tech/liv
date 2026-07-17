package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
	"github.com/fxkt-tech/liv/ffvmix"
)

const outputCount = 3

func main() {
	ctx := context.Background()
	baseDir := os.Getenv("FFVMIX_ASSET_DIR")
	if baseDir == "" {
		log.Fatal("FFVMIX_ASSET_DIR is required")
	}

	openingDuration, err := ffcut.NewDuration(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	fadeDuration, err := ffcut.NewDuration(500 * time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	cueDuration, err := ffcut.NewDuration(2 * time.Second)
	if err != nil {
		log.Fatal(err)
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

	// TargetDuration makes this slot exactly five seconds. The configured
	// policies center-trim long videos and loop short videos.
	opening, err := template.AddSlot(ffvmix.SlotConfig{
		Name:           "opening",
		TargetDuration: &openingDuration,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range []string{"video/opening-a.mp4", "video/opening-b.mp4"} {
		if _, err := opening.AddVideo(ffvmix.VideoSourceConfig{Path: path}); err != nil {
			log.Fatal(err)
		}
	}

	// Omitting TargetDuration keeps the selected video's natural duration.
	body, err := template.AddSlot(ffvmix.SlotConfig{Name: "body"})
	if err != nil {
		log.Fatal(err)
	}
	for _, path := range []string{"video/body-a.mp4", "video/body-b.mp4"} {
		if _, err := body.AddVideo(ffvmix.VideoSourceConfig{Path: path}); err != nil {
			log.Fatal(err)
		}
	}

	join, err := template.AddJoin(ffvmix.JoinConfig{
		FromSlotID: opening.ID,
		ToSlotID:   body.ID,
	})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind: ffcut.TransitionKindCut,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err := join.AddTransition(ffvmix.TransitionConfig{
		Kind:           ffcut.TransitionKindFade,
		Duration:       fadeDuration,
		AudioCrossfade: true,
	}); err != nil {
		log.Fatal(err)
	}

	bgmGain := 0.25
	for _, path := range []string{"audio/bgm-a.wav", "audio/bgm-b.wav"} {
		if _, err := template.AddBGM(ffvmix.BGMConfig{
			Path: path,
			Loop: true,
			Gain: &bgmGain,
		}); err != nil {
			log.Fatal(err)
		}
	}

	opacity := 0.9
	if _, err := template.AddImageLayer(ffvmix.ImageLayerConfig{
		Timing:  ffvmix.FullOutputLayerTiming(),
		Path:    "image/logo.png",
		Opacity: &opacity,
		Geometry: ffcut.Geometry{
			Anchor: ffcut.AnchorTopLeft,
			X:      ffcut.Length{Value: 85, Unit: ffcut.LengthUnitPercent},
			Y:      ffcut.Length{Value: 3, Unit: ffcut.LengthUnitPercent},
			Width:  ffcut.Length{Value: 12, Unit: ffcut.LengthUnitPercent},
			Height: ffcut.Length{Value: 12, Unit: ffcut.LengthUnitPercent},
		},
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := template.AddSubtitleLayer(ffvmix.SubtitleLayerConfig{
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
			Text:  "FFVMix example",
		}}),
	}); err != nil {
		log.Fatal(err)
	}

	template.AddMaxSimilarity(0.8)
	template.AddMaxVideoAssetUses(2)
	template.AddMaxBGMUses(2)

	templateJSON, err := ffvmix.MarshalTemplate(template)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("template:\n%s\n", templateJSON)

	// Compile owns all local file access and FFprobe work.
	compiled, err := ffvmix.Compile(ctx, template,
		ffvmix.WithBaseDir(baseDir),
		ffvmix.WithProbeConcurrency(4),
	)
	if err != nil {
		log.Fatal(err)
	}

	generator, err := ffvmix.NewGenerator(compiled,
		ffvmix.WithSeed(42),
		ffvmix.WithSearchBudget(100),
		ffvmix.WithConstraintFunc(
			"no-draft-assets",
			"no-draft-assets/v1",
			func(candidate ffvmix.CandidateView, _ ffvmix.HistoryView) (ffvmix.Decision, error) {
				for _, video := range candidate.Videos() {
					if strings.Contains(strings.ToLower(filepath.Base(video.Path)), "draft") {
						return ffvmix.Reject("draft_asset"), nil
					}
				}
				return ffvmix.Accept(), nil
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// FFVMix does not own an output count. The caller simply stops after taking
	// the number of results it needs.
	for generated := 0; generated < outputCount; {
		result, err := generator.Next(ctx)
		if err != nil {
			log.Fatal(err)
		}
		switch result.Status {
		case ffvmix.Yielded:
			projectJSON, err := ffcut.Marshal(result.Project)
			if err != nil {
				log.Fatal(err)
			}
			generated++
			fmt.Printf("project %d:\n%s\n", generated, projectJSON)
		case ffvmix.BudgetExceeded:
			continue
		case ffvmix.Exhausted:
			fmt.Printf("combination space exhausted after %d output(s)\n", generated)
			return
		}
	}
}
