package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
)

func TestRenderIntegration(t *testing.T) {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	ffprobeBin, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe is not installed")
	}

	directory := t.TempDir()
	firstPath := filepath.Join(directory, "red.mp4")
	secondPath := filepath.Join(directory, "blue.mp4")
	voicePath := filepath.Join(directory, "voice.wav")
	createVideoFixture(t, ffmpegBin, firstPath, "red", 440)
	createVideoFixture(t, ffmpegBin, secondPath, "blue", 550)
	runFixtureCommand(t, ffmpegBin,
		"-v", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=1000:duration=0.8",
		"-c:a", "pcm_s16le", voicePath,
	)

	project := integrationProject(t, firstPath, secondPath, voicePath)
	outputPath := filepath.Join(directory, "final.mp4")
	if err := Render(context.Background(), project, outputPath, WithFFmpegBin(ffmpegBin)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	probeOutput, err := exec.Command(
		ffprobeBin,
		"-v", "error",
		"-show_entries", "format=duration:stream=codec_type,width,height",
		"-of", "json",
		outputPath,
	).Output()
	if err != nil {
		t.Fatalf("ffprobe error = %v", err)
	}
	var probe struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(probeOutput, &probe); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var videoStreams, audioStreams int
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			videoStreams++
			if stream.Width != 180 || stream.Height != 320 {
				t.Errorf("video size = %dx%d, want 180x320", stream.Width, stream.Height)
			}
		case "audio":
			audioStreams++
		}
	}
	if videoStreams != 1 || audioStreams != 1 {
		t.Fatalf("streams = video:%d audio:%d, want 1 and 1", videoStreams, audioStreams)
	}
	var duration float64
	if _, err := fmt.Sscanf(probe.Format.Duration, "%f", &duration); err != nil {
		t.Fatalf("duration %q: %v", probe.Format.Duration, err)
	}
	if math.Abs(duration-0.8) > 0.08 {
		t.Errorf("duration = %.3f, want 0.8±0.08", duration)
	}

	assertFrameColor(t, ffmpegBin, outputPath, 0.1, 'r')
	assertFrameColor(t, ffmpegBin, outputPath, 0.6, 'b')
}

func createVideoFixture(t *testing.T, ffmpegBin, outputPath, color string, frequency int) {
	t.Helper()
	runFixtureCommand(t, ffmpegBin,
		"-v", "error", "-y",
		"-f", "lavfi", "-i", "color=c="+color+":s=320x240:r=30:d=0.4",
		"-f", "lavfi", "-i", fmt.Sprintf("sine=frequency=%d:duration=0.4", frequency),
		"-shortest", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", outputPath,
	)
}

func runFixtureCommand(t *testing.T, bin string, args ...string) {
	t.Helper()
	output, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v: %s", filepath.Base(bin), err, output)
	}
}

func integrationProject(t *testing.T, firstPath, secondPath, voicePath string) *ffcut.Project {
	t.Helper()
	clipDuration := 400 * time.Millisecond
	totalDuration := 2 * clipDuration
	return &ffcut.Project{
		Version: ffcut.ProjectVersion,
		ID:      "integration",
		Canvas: ffcut.Canvas{
			Width:     180,
			Height:    320,
			FrameRate: ffcut.FrameRate{Numerator: 30, Denominator: 1},
			Background: ffcut.Background{
				Kind:  ffcut.BackgroundKindColor,
				Color: &ffcut.ColorBackground{Color: "#000000"},
			},
		},
		Video: ffcut.Sequence{
			Clips: []ffcut.VideoClip{
				integrationClip(t, "red", firstPath, 0, clipDuration),
				integrationClip(t, "blue", secondPath, clipDuration, clipDuration),
			},
			Transitions: []ffcut.Transition{
				{
					ID:         "cut",
					Kind:       ffcut.TransitionKindCut,
					FromClipID: "red",
					ToClipID:   "blue",
					Range:      rendererRange(clipDuration, 0),
				},
			},
		},
		Audio: []ffcut.AudioTrack{
			{
				ID:            "voice",
				Kind:          ffcut.AudioTrackKindVoice,
				Source:        sourceFromFile(t, "voice-source", voicePath),
				SourceRange:   rendererRange(0, totalDuration),
				TimelineRange: rendererRange(0, totalDuration),
				Gain:          1,
			},
		},
		Metadata: ffcut.Metadata{
			TemplateFingerprint:    "integration-template",
			CombinationFingerprint: "integration-combination",
			Selections: []ffcut.Selection{
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-1", OptionID: "red", AssetFingerprint: "red"},
				{Kind: ffcut.SelectionKindVideo, DimensionID: "slot-2", OptionID: "blue", AssetFingerprint: "blue"},
				{Kind: ffcut.SelectionKindVoice, DimensionID: "voice", OptionID: "voice", AssetFingerprint: "voice"},
			},
		},
	}
}

func integrationClip(t *testing.T, id ffcut.ID, path string, start, duration time.Duration) ffcut.VideoClip {
	t.Helper()
	return ffcut.VideoClip{
		ID:            id,
		Source:        sourceFromFile(t, id+"-source", path),
		SourceRange:   rendererRange(0, duration),
		TimelineRange: rendererRange(start, duration),
		Playback:      ffcut.Playback{Rate: 1},
		Fit:           ffcut.FitModeCover,
		Audio:         ffcut.ClipAudio{},
	}
}

func sourceFromFile(t *testing.T, id ffcut.ID, path string) ffcut.LocalSource {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
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

func assertFrameColor(t *testing.T, ffmpegBin, path string, at float64, want byte) {
	t.Helper()
	output, err := exec.Command(
		ffmpegBin,
		"-v", "error",
		"-ss", fmt.Sprintf("%.3f", at),
		"-i", path,
		"-frames:v", "1",
		"-vf", "scale=1:1",
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-",
	).Output()
	if err != nil {
		t.Fatalf("extract frame at %.3f error = %v", at, err)
	}
	if len(output) < 3 {
		t.Fatalf("extract frame at %.3f returned %d bytes", at, len(output))
	}
	red, _, blue := output[0], output[1], output[2]
	switch want {
	case 'r':
		if red < 180 || red <= blue*2 {
			t.Errorf("frame at %.3f RGB = %v, want red", at, output[:3])
		}
	case 'b':
		if blue < 180 || blue <= red*2 {
			t.Errorf("frame at %.3f RGB = %v, want blue", at, output[:3])
		}
	}
}
