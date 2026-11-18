package fusion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type recordingTimelineRunner struct {
	bin    string
	args   []string
	output []byte
	err    error
}

func (r *recordingTimelineRunner) Run(_ context.Context, bin string, args []string) ([]byte, error) {
	r.bin = bin
	r.args = append([]string(nil), args...)
	return r.output, r.err
}

func TestTimelineValidateReportsPaths(t *testing.T) {
	timeline := validTimeline()
	timeline.Width = 0
	timeline.Tracks[0].Clips[0].Range.DurationMs = 0
	timeline.Tracks[1].Clips[0].Path = "relative.wav"
	timeline.Tracks[1].Clips[0].Audio = TimelineAudioConfig{Enabled: false, Volume: 1}

	err := timeline.Validate()
	if !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("Validate() error = %v, want ErrInvalidTimeline", err)
	}
	for _, path := range []string{
		"width",
		"tracks[0].clips[0].range.duration_ms",
		"tracks[1].clips[0].path",
		"tracks[1].clips[0].audio.volume",
	} {
		if !strings.Contains(err.Error(), path) {
			t.Errorf("Validate() error = %v, want path %q", err, path)
		}
	}
}

func TestTimelineValidateRejectsUnsupportedAndDuplicateItems(t *testing.T) {
	timeline := validTimeline()
	timeline.Tracks[1].ID = timeline.Tracks[0].ID
	timeline.Tracks[1].Clips[0].ID = timeline.Tracks[0].Clips[0].ID
	timeline.Tracks[0].Clips[0].Transform.Fit = "diagonal"
	timeline.Tracks[0].Clips[0].Transform.Crop = &TimelineCrop{X: 0.8, Y: 0, Width: 0.4, Height: 1}

	err := timeline.Validate()
	if !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("Validate() error = %v, want ErrInvalidTimeline", err)
	}
	for _, fragment := range []string{"duplicates", "diagonal", "normalized rectangle"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("Validate() error = %v, want %q", err, fragment)
		}
	}
}

func TestRenderUsesDeterministicVisualAndAudioOrderWithoutMutatingInput(t *testing.T) {
	timeline := validTimeline()
	bottom := timeline.Tracks[0]
	top := &TimelineTrack{
		ID: "a-top", Kind: TimelineTrackVisual, Order: 20,
		Clips: []*TimelineClip{
			visualClip("z-second", "/tmp/z-second.png", TimelineClipImage, 20, 0, 3000),
			visualClip("a-first", "/tmp/a-first.png", TimelineClipImage, 10, 0, 3000),
		},
	}
	bottom.Order = 10
	timeline.Tracks = []*TimelineTrack{top, timeline.Tracks[1], bottom}
	wantTrackIDs := []string{timeline.Tracks[0].ID, timeline.Tracks[1].ID, timeline.Tracks[2].ID}
	wantTopClipIDs := []string{top.Clips[0].ID, top.Clips[1].ID}
	runner := &recordingTimelineRunner{}

	if err := Render(context.Background(), timeline, "/tmp/final.mp4", withTimelineRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	command := strings.Join(runner.args, " ")
	paths := []string{"/tmp/video.mp4", "/tmp/a-first.png", "/tmp/z-second.png", "/tmp/audio.wav"}
	last := -1
	for _, path := range paths {
		index := strings.Index(command, "-i "+path)
		if index <= last {
			t.Fatalf("command order = %q, want %q after prior input", command, path)
		}
		last = index
	}
	gotTrackIDs := []string{timeline.Tracks[0].ID, timeline.Tracks[1].ID, timeline.Tracks[2].ID}
	gotTopClipIDs := []string{top.Clips[0].ID, top.Clips[1].ID}
	if !reflect.DeepEqual(gotTrackIDs, wantTrackIDs) || !reflect.DeepEqual(gotTopClipIDs, wantTopClipIDs) {
		t.Fatalf("Render() mutated caller ordering")
	}
}

func TestRenderBuildsFiniteTransformAudioAndLUTGraph(t *testing.T) {
	timeline := validTimeline()
	timeline.LUTFile = "/tmp/look,one.cube"
	clip := timeline.Tracks[0].Clips[0]
	clip.Transform = TimelineTransform{
		X: 20, Y: 30, Width: 160, Height: 90, RotationDegrees: 90, Opacity: 0.5,
		Fit: TimelineFitCover, Crop: &TimelineCrop{X: 0.1, Y: 0.2, Width: 0.8, Height: 0.6},
	}
	clip.HasAudio = true
	clip.Audio = TimelineAudioConfig{Enabled: true, Volume: 0.25}
	clip.SourceRange.DurationMs = 6000
	runner := &recordingTimelineRunner{}

	if err := Render(context.Background(), timeline, "/tmp/final.mp4", withTimelineRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	command := strings.Join(runner.args, " ")
	for _, want := range []string{
		"crop=iw*0.800000:ih*0.600000:iw*0.100000:ih*0.200000",
		"scale=160:90:force_original_aspect_ratio=increase",
		"colorchannelmixer=aa=0.500000",
		"rotate=90.000000*PI/180:c=none",
		"setpts=(PTS-STARTPTS)/2.000000",
		"repeatlast=0",
		"atempo=2.000000",
		"volume=0.250000",
		"lut3d=file='/tmp/look\\,one.cube'",
		"-r 25.000000",
		"-t 3.000000",
	} {
		if !strings.Contains(command, want) {
			t.Errorf("command = %q, want %q", command, want)
		}
	}
}

func TestRenderHonorsHiddenAndMutedTracks(t *testing.T) {
	timeline := validTimeline()
	timeline.Tracks[0].Hidden = true
	timeline.Tracks[1].Muted = true
	runner := &recordingTimelineRunner{}
	if err := Render(context.Background(), timeline, "/tmp/final.mp4", withTimelineRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	command := strings.Join(runner.args, " ")
	if strings.Contains(command, "/tmp/video.mp4") || strings.Contains(command, "/tmp/audio.wav") {
		t.Fatalf("command includes hidden/muted source: %q", command)
	}
	if !strings.Contains(command, "anullsrc") {
		t.Fatalf("command = %q, want finite silent audio", command)
	}
}

func TestRenderCanOmitAudioStream(t *testing.T) {
	timeline := validTimeline()
	timeline.OmitAudio = true
	runner := &recordingTimelineRunner{}
	if err := Render(context.Background(), timeline, "/tmp/final.mp4", withTimelineRunner(runner)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	command := strings.Join(runner.args, " ")
	for _, unwanted := range []string{"/tmp/audio.wav", "anullsrc", "[aout]", "-c:a"} {
		if strings.Contains(command, unwanted) {
			t.Fatalf("command = %q, must not contain %q", command, unwanted)
		}
	}
}

func TestRenderFailureAndCancellation(t *testing.T) {
	timeline := validTimeline()
	processErr := errors.New("exit status 1")
	err := Render(context.Background(), timeline, "/tmp/final.mp4", withTimelineRunner(&recordingTimelineRunner{
		output: []byte("bad filter"), err: processErr,
	}))
	if !errors.Is(err, ErrTimelineRenderFailed) || !errors.Is(err, processErr) || !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("Render() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = Render(ctx, timeline, "/tmp/final.mp4", withTimelineRunner(&recordingTimelineRunner{err: context.Canceled}))
	if !errors.Is(err, ErrTimelineRenderFailed) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Render() canceled error = %v", err)
	}
}

func TestTimelineRenderIntegration(t *testing.T) {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	ffprobeBin, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe is not installed")
	}
	directory := t.TempDir()
	videoPath := filepath.Join(directory, "silent.mp4")
	imagePath := filepath.Join(directory, "overlay.png")
	animationPath := filepath.Join(directory, "animation.gif")
	audioPath := filepath.Join(directory, "audio.wav")
	lutPath := filepath.Join(directory, "identity.cube")
	outputPath := filepath.Join(directory, "output.mp4")
	runTimelineFixture(t, ffmpegBin,
		"-f", "lavfi", "-i", "color=c=red:s=160x90:r=20:d=0.8",
		"-an", "-c:v", "libx264", "-pix_fmt", "yuv420p", videoPath,
	)
	runTimelineFixture(t, ffmpegBin,
		"-f", "lavfi", "-i", "color=c=yellow:s=160x90",
		"-frames:v", "1", imagePath,
	)
	runTimelineFixture(t, ffmpegBin,
		"-f", "lavfi", "-i", "color=c=lime:s=160x90:r=10:d=0.1",
		"-f", "lavfi", "-i", "color=c=magenta:s=160x90:r=10:d=0.1",
		"-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0,split[x][p];[p]palettegen[pal];[x][pal]paletteuse",
		"-loop", "0", animationPath,
	)
	runTimelineFixture(t, ffmpegBin,
		"-f", "lavfi", "-i", "sine=frequency=440:duration=0.8",
		"-c:a", "pcm_s16le", audioPath,
	)
	if err := os.WriteFile(lutPath, []byte(strings.TrimSpace(`
TITLE "identity"
LUT_3D_SIZE 2
DOMAIN_MIN 0.0 0.0 0.0
DOMAIN_MAX 1.0 1.0 1.0
0.0 0.0 0.0
1.0 0.0 0.0
0.0 1.0 0.0
1.0 1.0 0.0
0.0 0.0 1.0
1.0 0.0 1.0
0.0 1.0 1.0
1.0 1.0 1.0
`)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	timeline := &Timeline{
		Width: 160, Height: 90, FPS: 20, Background: "black", DurationMs: 800, LUTFile: lutPath,
		Tracks: []*TimelineTrack{
			{ID: "base", Kind: TimelineTrackVisual, Order: 0, Clips: []*TimelineClip{
				visualClip("video", videoPath, TimelineClipVideo, 0, 0, 800),
			}},
			{ID: "overlays", Kind: TimelineTrackVisual, Order: 10, Clips: []*TimelineClip{
				visualClip("image", imagePath, TimelineClipImage, 0, 0, 200),
				visualClip("animation", animationPath, TimelineClipAnimation, 10, 200, 400),
			}},
			{ID: "sound", Kind: TimelineTrackAudio, Order: 0, Clips: []*TimelineClip{
				{
					ID: "audio", Kind: TimelineClipAudio, Path: audioPath,
					Range: TimelineRange{DurationMs: 800}, SourceRange: TimelineRange{DurationMs: 800},
					Audio: TimelineAudioConfig{Enabled: true, Volume: 0.5},
				},
			}},
		},
	}
	for _, track := range timeline.Tracks[:2] {
		for _, clip := range track.Clips {
			clip.Transform.Width = 160
			clip.Transform.Height = 90
		}
	}
	if err := Render(context.Background(), timeline, outputPath, WithTimelineFFmpegBin(ffmpegBin)); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	probeOutput, err := exec.Command(ffprobeBin,
		"-v", "error", "-show_entries", "format=duration:stream=codec_type,width,height,avg_frame_rate", "-of", "json", outputPath,
	).Output()
	if err != nil {
		t.Fatalf("ffprobe error = %v", err)
	}
	var probe struct {
		Streams []struct {
			CodecType    string `json:"codec_type"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			AvgFrameRate string `json:"avg_frame_rate"`
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
			if stream.Width != 160 || stream.Height != 90 || stream.AvgFrameRate != "20/1" {
				t.Errorf("video stream = %#v", stream)
			}
		case "audio":
			audioStreams++
		}
	}
	if videoStreams != 1 || audioStreams != 1 {
		t.Fatalf("streams = video:%d audio:%d", videoStreams, audioStreams)
	}
	var duration float64
	if _, err := fmt.Sscanf(probe.Format.Duration, "%f", &duration); err != nil {
		t.Fatalf("duration %q: %v", probe.Format.Duration, err)
	}
	if math.Abs(duration-0.8) > 0.08 {
		t.Errorf("duration = %.3f, want 0.8±0.08", duration)
	}
	assertTimelineFrameColor(t, ffmpegBin, outputPath, 0.1, 'y')
	assertTimelineFrameColor(t, ffmpegBin, outputPath, 0.7, 'r')
}

func validTimeline() *Timeline {
	return &Timeline{
		Width: 1920, Height: 1080, FPS: 25, Background: "#000000", DurationMs: 3000,
		Tracks: []*TimelineTrack{
			{
				ID: "video-track", Kind: TimelineTrackVisual, Order: 0,
				Clips: []*TimelineClip{visualClip("video", "/tmp/video.mp4", TimelineClipVideo, 0, 0, 3000)},
			},
			{
				ID: "audio-track", Kind: TimelineTrackAudio, Order: 0,
				Clips: []*TimelineClip{{
					ID: "audio", Kind: TimelineClipAudio, Path: "/tmp/audio.wav",
					Range: TimelineRange{DurationMs: 3000}, SourceRange: TimelineRange{DurationMs: 3000},
					Audio: TimelineAudioConfig{Enabled: true, Volume: 1},
				}},
			},
		},
	}
}

func visualClip(id, path string, kind TimelineClipKind, order int, startMs, durationMs int64) *TimelineClip {
	clip := &TimelineClip{
		ID: id, Kind: kind, Order: order, Path: path,
		Range:     TimelineRange{StartMs: startMs, DurationMs: durationMs},
		Transform: TimelineTransform{Width: 1920, Height: 1080, Opacity: 1, Fit: TimelineFitStretch},
	}
	if kind == TimelineClipVideo {
		clip.SourceRange = TimelineRange{DurationMs: durationMs}
	}
	if kind == TimelineClipAnimation {
		clip.Loop = true
	}
	return clip
}

func runTimelineFixture(t *testing.T, bin string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-v", "error", "-y"}, args...)
	output, err := exec.Command(bin, commandArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("ffmpeg fixture failed: %v: %s", err, output)
	}
}

func assertTimelineFrameColor(t *testing.T, ffmpegBin, videoPath string, second float64, want byte) {
	t.Helper()
	outputPath := filepath.Join(t.TempDir(), "pixel.ppm")
	output, err := exec.Command(ffmpegBin,
		"-v", "error", "-y", "-ss", fmt.Sprintf("%.3f", second), "-i", videoPath,
		"-frames:v", "1", "-vf", "scale=1:1", outputPath,
	).CombinedOutput()
	if err != nil {
		t.Fatalf("extract frame failed: %v: %s", err, output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	newline := strings.Index(string(data), "\n255\n")
	if newline < 0 || newline+5+3 > len(data) {
		t.Fatalf("unexpected PPM fixture")
	}
	pixel := data[newline+5 : newline+8]
	var got byte
	switch {
	case pixel[0] > 150 && pixel[1] > 150 && pixel[2] < 100:
		got = 'y'
	case pixel[0] > pixel[1]+40 && pixel[0] > pixel[2]+40:
		got = 'r'
	default:
		got = '?'
	}
	if got != want {
		t.Fatalf("frame %.3f pixel = %v (%c), want %c", second, pixel, got, want)
	}
}
