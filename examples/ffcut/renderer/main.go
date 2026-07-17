package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
	ffcutrenderer "github.com/fxkt-tech/liv/ffcut/renderer"
)

const (
	defaultWidth     = 720
	defaultHeight    = 1280
	defaultFrameRate = 30
	defaultDuration  = 4 * time.Second
)

type options struct {
	clipA    string
	clipB    string
	voice    string
	output   string
	duration time.Duration
	width    int
	height   int
	debug    bool
}

func main() {
	config := options{}
	flag.StringVar(&config.clipA, "clip-a", "", "first local video path")
	flag.StringVar(&config.clipB, "clip-b", "", "second local video path")
	flag.StringVar(&config.voice, "voice", "", "full-timeline narration audio path")
	flag.StringVar(&config.output, "out", "final.mp4", "output MP4 path")
	flag.DurationVar(&config.duration, "duration", defaultDuration, "output duration split evenly across both clips")
	flag.IntVar(&config.width, "width", defaultWidth, "output width")
	flag.IntVar(&config.height, "height", defaultHeight, "output height")
	flag.BoolVar(&config.debug, "debug", false, "print the quoted FFmpeg command")
	flag.Parse()

	if err := run(context.Background(), config); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, config options) error {
	if config.clipA == "" || config.clipB == "" || config.voice == "" {
		return fmt.Errorf("clip-a, clip-b, and voice are required")
	}
	if config.duration <= 0 || config.duration%(2*time.Microsecond) != 0 {
		return fmt.Errorf("duration must be positive and divisible by two microseconds")
	}
	if config.width <= 0 || config.height <= 0 {
		return fmt.Errorf("width and height must be positive")
	}

	clipASource, clipAFingerprint, err := localSource("clip-a-source", config.clipA)
	if err != nil {
		return err
	}
	clipBSource, clipBFingerprint, err := localSource("clip-b-source", config.clipB)
	if err != nil {
		return err
	}
	voiceSource, voiceFingerprint, err := localSource("voice-source", config.voice)
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs(config.output)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	clipDuration := config.duration / 2
	projectClipDuration, err := ffcut.NewDuration(clipDuration)
	if err != nil {
		return fmt.Errorf("convert clip duration: %w", err)
	}
	projectDuration, err := ffcut.NewDuration(config.duration)
	if err != nil {
		return fmt.Errorf("convert project duration: %w", err)
	}
	project := &ffcut.Project{
		Version: ffcut.ProjectVersion,
		ID:      "renderer-example",
		Canvas: ffcut.Canvas{
			Width:     int32(config.width),
			Height:    int32(config.height),
			FrameRate: ffcut.FrameRate{Numerator: defaultFrameRate, Denominator: 1},
			Background: ffcut.Background{
				Kind:  ffcut.BackgroundKindColor,
				Color: &ffcut.ColorBackground{Color: "#000000"},
			},
		},
		Video: ffcut.Sequence{
			Clips: []ffcut.VideoClip{
				videoClip("clip-a", clipASource, 0, projectClipDuration),
				videoClip("clip-b", clipBSource, projectClipDuration, projectClipDuration),
			},
			Transitions: []ffcut.Transition{{
				ID:         "cut-a-b",
				Kind:       ffcut.TransitionKindCut,
				FromClipID: "clip-a",
				ToClipID:   "clip-b",
				Range:      ffcut.TimeRange{Start: projectClipDuration},
			}},
		},
		Audio: []ffcut.AudioTrack{{
			ID:            "voice-track",
			Kind:          ffcut.AudioTrackKindVoice,
			Source:        voiceSource,
			SourceRange:   ffcut.TimeRange{Duration: projectDuration},
			TimelineRange: ffcut.TimeRange{Duration: projectDuration},
			Gain:          1,
		}},
		Metadata: ffcut.Metadata{
			TemplateFingerprint:    "renderer-example/v1",
			CombinationFingerprint: "clip-a+clip-b+voice",
			Selections: []ffcut.Selection{
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-a", OptionID: "clip-a", AssetFingerprint: clipAFingerprint},
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-b", OptionID: "clip-b", AssetFingerprint: clipBFingerprint},
				{Kind: ffcut.SelectionKindVoice, DimensionID: "voice", OptionID: "narration", AssetFingerprint: voiceFingerprint},
			},
		},
	}

	projectJSON, err := ffcut.Marshal(project)
	if err != nil {
		return fmt.Errorf("validate project: %w", err)
	}
	fmt.Printf("validated project:\n%s\n", projectJSON)

	renderOptions := make([]ffcutrenderer.Option, 0, 1)
	if config.debug {
		renderOptions = append(renderOptions, ffcutrenderer.WithDebug(os.Stdout))
	}
	if err := ffcutrenderer.Render(ctx, project, outputPath, renderOptions...); err != nil {
		return fmt.Errorf("render project: %w", err)
	}
	fmt.Printf("rendered MP4: %s\n", outputPath)
	return nil
}

func videoClip(id ffcut.ID, source ffcut.LocalSource, start, duration ffcut.Duration) ffcut.VideoClip {
	return ffcut.VideoClip{
		ID:            id,
		Source:        source,
		SourceRange:   ffcut.TimeRange{Duration: duration},
		TimelineRange: ffcut.TimeRange{Start: start, Duration: duration},
		Playback:      ffcut.Playback{Rate: 1},
		Fit:           ffcut.FitModeCover,
		Audio:         ffcut.ClipAudio{Enabled: false},
	}
}

func localSource(id ffcut.ID, path string) (ffcut.LocalSource, string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return ffcut.LocalSource{}, "", fmt.Errorf("resolve %s: %w", id, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return ffcut.LocalSource{}, "", fmt.Errorf("stat %s: %w", id, err)
	}
	fingerprint := fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
	return ffcut.LocalSource{
		ID:   id,
		Path: absolutePath,
		Fingerprint: ffcut.MediaFingerprint{
			Size:             info.Size(),
			ModifiedUnixNano: info.ModTime().UnixNano(),
		},
	}, fingerprint, nil
}
