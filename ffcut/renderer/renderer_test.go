package renderer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

type recordingRunner struct {
	bin  string
	args []string
	err  error
}

func (runner *recordingRunner) Run(_ context.Context, bin string, args []string) ([]byte, error) {
	runner.bin = bin
	runner.args = append([]string(nil), args...)
	if runner.err != nil {
		return []byte("render detail"), runner.err
	}
	return nil, nil
}

func TestRenderBuildsVMixCommand(t *testing.T) {
	project := vmixProject(t)
	runner := &recordingRunner{}
	output := filepath.Join(t.TempDir(), "final.mp4")

	if err := Render(context.Background(), project, output, withRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if runner.bin != defaultFFmpegBin {
		t.Fatalf("bin = %q, want %q", runner.bin, defaultFFmpegBin)
	}
	command := strings.Join(runner.args, " ")
	for _, want := range []string{
		"scale=720:1280:force_original_aspect_ratio=increase",
		"crop=720:1280:(iw-ow)/2:(ih-oh)/2",
		"concat=n=2:v=1:a=0[vout]",
		"[2:a:0]atrim=duration=2.000000",
		"-map [vout] -map [aout]",
		"-c:v libx264",
		"-c:a aac",
		"-movflags +faststart",
	} {
		if !strings.Contains(command, want) {
			t.Errorf("command = %q, want fragment %q", command, want)
		}
	}
	if strings.Contains(command, "[0:a") || strings.Contains(command, "[1:a") {
		t.Fatalf("command maps original clip audio: %q", command)
	}
}

func TestRenderBuildsOrderedLayerGraph(t *testing.T) {
	project := vmixProject(t)
	directory := t.TempDir()
	textOpacity := 0.55
	image := rendererSource(t, "image", filepath.Join(directory, "sticker.png"))
	animation := rendererSource(t, "animation", filepath.Join(directory, "sticker.gif"))
	project.Layers = []ffcut.Layer{
		{
			ID: "image-layer", Kind: ffcut.LayerKindImage, Range: rendererRange(0, 2*time.Second),
			Image: &ffcut.ImageLayer{
				Source: image, Geometry: rendererGeometry(10, 20, 100, 80), Opacity: 0.8,
			},
		},
		{
			ID: "animation-layer", Kind: ffcut.LayerKindMedia, Range: rendererRange(0, 2*time.Second),
			Media: &ffcut.MediaLayer{
				Kind: ffcut.MediaKindAnimation, Source: animation,
				Geometry: rendererGeometry(120, 40, 90, 90), Opacity: 1, Loop: true,
			},
		},
		{
			ID: "text-layer", Kind: ffcut.LayerKindSubtitle, Range: rendererRange(0, 2*time.Second),
			Subtitle: &ffcut.SubtitleLayer{
				Region:          rendererGeometry(40, 100, 300, 120),
				Opacity:         &textOpacity,
				RotationDegrees: 30,
				Style: ffcut.SubtitleStyle{
					FontFamily: "sans-serif", FontSize: ffcut.Length{Value: 42, Unit: ffcut.LengthUnitPixel},
					Color: "#FFFFFF", Align: ffcut.TextAlignCenter,
					StrokeColor: "#000000", StrokeWidth: ffcut.Length{Value: 2, Unit: ffcut.LengthUnitPixel},
				},
				Cues: []ffcut.SubtitleCue{{ID: "title", Range: rendererRange(0, 2*time.Second), Text: "标题"}},
			},
		},
	}
	runner := &recordingRunner{}
	if err := Render(context.Background(), project, filepath.Join(directory, "final.mp4"), withRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	command := strings.Join(runner.args, " ")
	for _, want := range []string{
		"-loop 1 -t 2.000000 -i " + image.Path,
		"-ignore_loop 0 -stream_loop -1 -t 2.000000 -i " + animation.Path,
		"colorchannelmixer=aa=0.800000",
		"overlay=x=10:y=20",
		"drawtext=textfile='",
		"expansion=none",
		"borderw=2.000000:bordercolor=#000000",
		"colorchannelmixer=aa=0.550000",
		"rotate=30.000000*PI/180:c=none",
		"overlay=x=30:y=33",
		"[4:a:0]atrim=duration=2.000000",
	} {
		if !strings.Contains(command, want) {
			t.Errorf("command = %q, want fragment %q", command, want)
		}
	}
}

func TestRotatedOverlayPositionKeepsOriginalCenter(t *testing.T) {
	canvas := ffcut.Canvas{Width: 720, Height: 1280}
	x, y := rotatedOverlayPosition(rendererGeometry(10, 20, 100, 80), 90, canvas)
	if x != 20 || y != 10 {
		t.Fatalf("position = (%d, %d), want (20, 10)", x, y)
	}
}

func TestRenderRejectsUnsupportedFeatures(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ffcut.Project)
		path string
	}{
		{
			name: "clip audio",
			path: "video.clips[0].audio.enabled",
			edit: func(project *ffcut.Project) {
				project.Video.Clips[0].Audio = ffcut.ClipAudio{Enabled: true, Gain: 1}
			},
		},
		{
			name: "fade",
			path: "video.transitions[0].kind",
			edit: func(project *ffcut.Project) {
				project.Video.Clips[0].SourceRange.Duration = rendererRange(0, 2*time.Second).Duration
				project.Video.Clips[0].TimelineRange.Duration = rendererRange(0, 2*time.Second).Duration
				project.Video.Clips[1].TimelineRange.Start = rendererRange(time.Second, time.Second).Start
				project.Video.Transitions[0].Kind = ffcut.TransitionKindFade
				project.Video.Transitions[0].Range = rendererRange(time.Second, time.Second)
			},
		},
		{
			name: "bgm instead of voice",
			path: "audio",
			edit: func(project *ffcut.Project) {
				project.Audio[0].Kind = ffcut.AudioTrackKindBGM
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := vmixProject(t)
			test.edit(project)
			err := Render(context.Background(), project, filepath.Join(t.TempDir(), "final.mp4"), withRunner(&recordingRunner{}))
			if !errors.Is(err, ErrUnsupportedProject) {
				t.Fatalf("Render() error = %v, want ErrUnsupportedProject", err)
			}
			if !strings.Contains(err.Error(), test.path) {
				t.Fatalf("Render() error = %v, want path %q", err, test.path)
			}
		})
	}
}

func TestRenderRejectsInvalidOutput(t *testing.T) {
	err := Render(context.Background(), vmixProject(t), "relative.mp4", withRunner(&recordingRunner{}))
	if !errors.Is(err, ErrInvalidOutput) {
		t.Fatalf("Render() error = %v, want ErrInvalidOutput", err)
	}
}

func TestRenderPreservesProcessFailure(t *testing.T) {
	processErr := errors.New("exit status 1")
	err := Render(
		context.Background(),
		vmixProject(t),
		filepath.Join(t.TempDir(), "final.mp4"),
		withRunner(&recordingRunner{err: processErr}),
	)
	if !errors.Is(err, ErrRenderFailed) || !errors.Is(err, processErr) {
		t.Fatalf("Render() error = %v, want render and process errors", err)
	}
	if !strings.Contains(err.Error(), "render detail") {
		t.Fatalf("Render() error = %v, want stderr", err)
	}
}

func vmixProject(t *testing.T) *ffcut.Project {
	t.Helper()
	directory := t.TempDir()
	first := rendererSource(t, "first", filepath.Join(directory, "first.mp4"))
	second := rendererSource(t, "second", filepath.Join(directory, "second.mp4"))
	voice := rendererSource(t, "voice", filepath.Join(directory, "voice.wav"))

	return &ffcut.Project{
		Version: ffcut.ProjectVersion,
		ID:      "vmix-output",
		Canvas: ffcut.Canvas{
			Width:     720,
			Height:    1280,
			FrameRate: ffcut.FrameRate{Numerator: 30, Denominator: 1},
			Background: ffcut.Background{
				Kind:  ffcut.BackgroundKindColor,
				Color: &ffcut.ColorBackground{Color: "#000000"},
			},
		},
		Video: ffcut.Sequence{
			Clips: []ffcut.VideoClip{
				{
					ID:            "clip-1",
					Source:        first,
					SourceRange:   rendererRange(0, time.Second),
					TimelineRange: rendererRange(0, time.Second),
					Playback:      ffcut.Playback{Rate: 1},
					Fit:           ffcut.FitModeCover,
					Audio:         ffcut.ClipAudio{},
				},
				{
					ID:            "clip-2",
					Source:        second,
					SourceRange:   rendererRange(0, time.Second),
					TimelineRange: rendererRange(time.Second, time.Second),
					Playback:      ffcut.Playback{Rate: 1},
					Fit:           ffcut.FitModeCover,
					Audio:         ffcut.ClipAudio{},
				},
			},
			Transitions: []ffcut.Transition{
				{
					ID:         "cut-1-2",
					Kind:       ffcut.TransitionKindCut,
					FromClipID: "clip-1",
					ToClipID:   "clip-2",
					Range:      rendererRange(time.Second, 0),
				},
			},
		},
		Audio: []ffcut.AudioTrack{
			{
				ID:            "voice-track",
				Kind:          ffcut.AudioTrackKindVoice,
				Source:        voice,
				SourceRange:   rendererRange(0, 2*time.Second),
				TimelineRange: rendererRange(0, 2*time.Second),
				Gain:          1,
			},
		},
		Metadata: ffcut.Metadata{
			TemplateFingerprint:    "template",
			CombinationFingerprint: "combination",
			Selections: []ffcut.Selection{
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-1", OptionID: "first", AssetFingerprint: "first"},
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-2", OptionID: "second", AssetFingerprint: "second"},
				{Kind: ffcut.SelectionKindVoice, DimensionID: "voice", OptionID: "voice", AssetFingerprint: "voice"},
			},
		},
	}
}

func rendererSource(t *testing.T, id ffcut.ID, path string) ffcut.LocalSource {
	t.Helper()
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	return ffcut.LocalSource{
		ID:   id,
		Path: path,
		Fingerprint: ffcut.MediaFingerprint{
			Size:             info.Size(),
			ModifiedUnixNano: info.ModTime().UnixNano(),
		},
	}
}

func rendererRange(start, duration time.Duration) ffcut.TimeRange {
	parsedStart, err := ffcut.NewDuration(start)
	if err != nil {
		panic(err)
	}
	parsedDuration, err := ffcut.NewDuration(duration)
	if err != nil {
		panic(err)
	}
	return ffcut.TimeRange{Start: parsedStart, Duration: parsedDuration}
}

func rendererGeometry(x, y, width, height float64) ffcut.Geometry {
	return ffcut.Geometry{
		Anchor: ffcut.AnchorTopLeft,
		X:      ffcut.Length{Value: x, Unit: ffcut.LengthUnitPixel},
		Y:      ffcut.Length{Value: y, Unit: ffcut.LengthUnitPixel},
		Width:  ffcut.Length{Value: width, Unit: ffcut.LengthUnitPixel},
		Height: ffcut.Length{Value: height, Unit: ffcut.LengthUnitPixel},
	}
}
