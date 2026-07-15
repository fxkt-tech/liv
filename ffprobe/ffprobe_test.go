package ffprobe_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fxkt-tech/liv/ffprobe"
)

func TestExtract(t *testing.T) {
	infile := generateProbeFixture(t)

	fp, err := ffprobe.New().Input(infile).Extract(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	vStream := fp.GetFirstVideoStream()
	if vStream == nil {
		t.Fatal("file has no v stream")
	}
	if vStream.Width != 16 || vStream.Height != 16 || vStream.Duration <= 0 {
		t.Fatalf("video stream = %#v, want 16x16 with positive duration", vStream)
	}
}

func TestProbeRunRaw(t *testing.T) {
	path := generateProbeFixture(t)
	probe := ffprobe.New().Input(path)
	raw, err := probe.RunRetRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"streams"`) {
		t.Fatalf("raw probe result has no streams: %s", raw)
	}
}

func generateProbeFixture(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe is not installed")
	}
	path := filepath.Join(t.TempDir(), "probe.mp4")
	command := exec.Command(
		"ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=black:s=16x16:r=10:d=0.2",
		"-c:v", "mpeg4", "-an", path,
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generate probe fixture: %v: %s", err, output)
	}
	return path
}
