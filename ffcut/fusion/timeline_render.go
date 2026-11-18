package fusion

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	ErrInvalidTimelineOutput = errors.New("invalid fusion timeline output")
	ErrTimelineRenderFailed  = errors.New("fusion timeline render failed")
)

type timelineCommandRunner interface {
	Run(context.Context, string, []string) ([]byte, error)
}

type execTimelineCommandRunner struct{}

func (execTimelineCommandRunner) Run(ctx context.Context, bin string, args []string) ([]byte, error) {
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

type TimelineRenderOption func(*timelineRenderConfig)

type timelineRenderConfig struct {
	ffmpegBin   string
	debugWriter io.Writer
	runner      timelineCommandRunner
}

// WithTimelineFFmpegBin selects the FFmpeg executable used by Render.
func WithTimelineFFmpegBin(bin string) TimelineRenderOption {
	return func(config *timelineRenderConfig) {
		if strings.TrimSpace(bin) != "" {
			config.ffmpegBin = bin
		}
	}
}

// WithTimelineDebug writes a safely quoted command before executing it.
func WithTimelineDebug(writer io.Writer) TimelineRenderOption {
	return func(config *timelineRenderConfig) {
		config.debugWriter = writer
	}
}

func withTimelineRunner(runner timelineCommandRunner) TimelineRenderOption {
	return func(config *timelineRenderConfig) {
		if runner != nil {
			config.runner = runner
		}
	}
}

// Render validates and renders a deterministic, finite timeline to MP4.
func Render(ctx context.Context, timeline *Timeline, outputPath string, opts ...TimelineRenderOption) error {
	if err := timeline.Validate(); err != nil {
		return err
	}
	if outputPath == "" || !filepath.IsAbs(outputPath) {
		return fmt.Errorf("%w: path must be absolute", ErrInvalidTimelineOutput)
	}
	config := timelineRenderConfig{ffmpegBin: "ffmpeg", runner: execTimelineCommandRunner{}}
	for _, option := range opts {
		if option != nil {
			option(&config)
		}
	}
	args := buildTimelineArgs(timeline, outputPath)
	if config.debugWriter != nil {
		writeTimelineCommand(config.debugWriter, config.ffmpegBin, args)
	}
	output, err := config.runner.Run(ctx, config.ffmpegBin, args)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%w: %w", ErrTimelineRenderFailed, ctx.Err())
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("%w: %w", ErrTimelineRenderFailed, err)
	}
	return fmt.Errorf("%w: %w: %s", ErrTimelineRenderFailed, err, detail)
}

type timelineInput struct {
	ordered orderedTimelineClip
	index   int
}

func buildTimelineArgs(timeline *Timeline, outputPath string) []string {
	visual := timeline.orderedClips(TimelineTrackVisual)
	audioTracks := timeline.orderedClips(TimelineTrackAudio)
	args := []string{"-v", "error", "-y"}
	visualInputs := make([]timelineInput, 0, len(visual))
	audioInputs := make([]timelineInput, 0, len(audioTracks))
	nextInput := 0
	for _, item := range visual {
		args = append(args, timelineInputArgs(item.clip, timeline.FPS)...)
		visualInputs = append(visualInputs, timelineInput{ordered: item, index: nextInput})
		nextInput++
	}
	if !timeline.OmitAudio {
		for _, item := range audioTracks {
			if item.track.Muted || !item.clip.Audio.Enabled {
				continue
			}
			args = append(args, timelineInputArgs(item.clip, timeline.FPS)...)
			audioInputs = append(audioInputs, timelineInput{ordered: item, index: nextInput})
			nextInput++
		}
	}

	duration := timelineSeconds(timeline.DurationMs)
	frameRate := timelineFloat(timeline.FPS)
	graph := []string{
		fmt.Sprintf("color=c=%s:s=%dx%d:r=%s:d=%s,format=rgba[canvas]", timeline.Background, timeline.Width, timeline.Height, frameRate, duration),
	}
	currentVideo := "canvas"
	for index, item := range visualInputs {
		clipLabel := fmt.Sprintf("visual%d", index)
		nextLabel := fmt.Sprintf("stage%d", index)
		graph = append(graph, timelineVisualFilter(item.index, clipLabel, timeline, item.ordered.clip))
		graph = append(graph, timelineOverlayFilter(currentVideo, clipLabel, nextLabel, item.ordered.clip))
		currentVideo = nextLabel
	}
	if timeline.LUTFile != "" {
		graph = append(graph, fmt.Sprintf("[%s]lut3d=file='%s':interp=tetrahedral[lutout]", currentVideo, escapeTimelineFilterPath(timeline.LUTFile)))
		currentVideo = "lutout"
	}
	graph = append(graph, fmt.Sprintf("[%s]format=yuv420p[vout]", currentVideo))

	if !timeline.OmitAudio {
		graph = append(graph, fmt.Sprintf("anullsrc=r=48000:cl=stereo:d=%s[asilence]", duration))
		audioLabels := []string{"[asilence]"}
		for _, item := range visualInputs {
			clip := item.ordered.clip
			if item.ordered.track.Muted || clip.Kind != TimelineClipVideo || !clip.HasAudio || !clip.Audio.Enabled {
				continue
			}
			label := fmt.Sprintf("clipaudio%d", len(audioLabels))
			graph = append(graph, timelineAudioFilter(item.index, label, clip, clip.Audio.Volume))
			audioLabels = append(audioLabels, "["+label+"]")
		}
		for _, item := range audioInputs {
			clip := item.ordered.clip
			label := fmt.Sprintf("trackaudio%d", len(audioLabels))
			graph = append(graph, timelineAudioFilter(item.index, label, clip, clip.Audio.Volume))
			audioLabels = append(audioLabels, "["+label+"]")
		}
		if len(audioLabels) == 1 {
			graph = append(graph, "[asilence]anull[aout]")
		} else {
			graph = append(graph, strings.Join(audioLabels, "")+fmt.Sprintf("amix=inputs=%d:duration=first:dropout_transition=0:normalize=0[aout]", len(audioLabels)))
		}
	}

	args = append(args,
		"-filter_complex", strings.Join(graph, ";"),
		"-map", "[vout]",
	)
	if !timeline.OmitAudio {
		args = append(args, "-map", "[aout]")
	}
	args = append(args,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-r", frameRate,
	)
	if !timeline.OmitAudio {
		args = append(args, "-c:a", "aac", "-ar", "48000", "-b:a", "192k")
	}
	args = append(args, "-movflags", "+faststart", "-t", duration, outputPath)
	return args
}

func timelineInputArgs(clip *TimelineClip, fps float64) []string {
	args := make([]string, 0, 12)
	switch clip.Kind {
	case TimelineClipImage:
		args = append(args, "-loop", "1", "-framerate", timelineFloat(fps), "-t", timelineSeconds(clip.Range.DurationMs))
	case TimelineClipAnimation:
		args = append(args, "-ignore_loop", "0", "-stream_loop", "-1", "-t", timelineSeconds(clip.Range.DurationMs))
	case TimelineClipVideo, TimelineClipAudio:
		if clip.Loop {
			args = append(args, "-stream_loop", "-1")
		}
		if clip.SourceRange.StartMs > 0 {
			args = append(args, "-ss", timelineSeconds(clip.SourceRange.StartMs))
		}
		args = append(args, "-t", timelineSeconds(clip.SourceRange.DurationMs))
	}
	return append(args, "-i", clip.Path)
}

func timelineVisualFilter(input int, label string, timeline *Timeline, clip *TimelineClip) string {
	transform := clip.Transform
	filters := make([]string, 0, 12)
	filters = append(filters, "format=rgba")
	if transform.Crop != nil {
		crop := transform.Crop
		filters = append(filters, fmt.Sprintf(
			"crop=iw*%s:ih*%s:iw*%s:ih*%s",
			timelineFloat(crop.Width), timelineFloat(crop.Height), timelineFloat(crop.X), timelineFloat(crop.Y),
		))
	}
	switch transform.Fit {
	case TimelineFitCover:
		filters = append(filters,
			fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase", transform.Width, transform.Height),
			fmt.Sprintf("crop=%d:%d:(iw-ow)/2:(ih-oh)/2", transform.Width, transform.Height),
		)
	case TimelineFitContain:
		filters = append(filters,
			fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", transform.Width, transform.Height),
			fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black@0", transform.Width, transform.Height),
		)
	case TimelineFitStretch:
		filters = append(filters, fmt.Sprintf("scale=%d:%d", transform.Width, transform.Height))
	}
	filters = append(filters, "setsar=1", "fps="+timelineFloat(timeline.FPS))
	if transform.Opacity != 1 {
		filters = append(filters, "colorchannelmixer=aa="+timelineFloat(transform.Opacity))
	}
	if transform.RotationDegrees != 0 {
		angle := timelineFloat(transform.RotationDegrees) + "*PI/180"
		filters = append(filters, fmt.Sprintf("rotate=%s:c=none:ow=rotw(%s):oh=roth(%s)", angle, angle, angle))
	}
	if clip.Kind == TimelineClipVideo && clip.SourceRange.DurationMs != clip.Range.DurationMs {
		rate := float64(clip.SourceRange.DurationMs) / float64(clip.Range.DurationMs)
		filters = append(filters, "setpts=(PTS-STARTPTS)/"+timelineFloat(rate))
	} else {
		filters = append(filters, "setpts=PTS-STARTPTS")
	}
	filters = append(filters,
		"trim=duration="+timelineSeconds(clip.Range.DurationMs),
		"setpts=PTS+"+timelineSeconds(clip.Range.StartMs)+"/TB",
	)
	return fmt.Sprintf("[%d:v:0]%s[%s]", input, strings.Join(filters, ","), label)
}

func timelineOverlayFilter(base, overlay, output string, clip *TimelineClip) string {
	x, y := timelineRotatedPosition(clip.Transform)
	end := clip.Range.StartMs + clip.Range.DurationMs
	enable := fmt.Sprintf("gte(t\\,%s)*lt(t\\,%s)", timelineSeconds(clip.Range.StartMs), timelineSeconds(end))
	return fmt.Sprintf(
		"[%s][%s]overlay=x=%d:y=%d:eof_action=pass:repeatlast=0:shortest=0:enable='%s'[%s]",
		base, overlay, x, y, enable, output,
	)
}

func timelineAudioFilter(input int, label string, clip *TimelineClip, volume float64) string {
	filters := []string{"asetpts=PTS-STARTPTS"}
	if clip.SourceRange.DurationMs != clip.Range.DurationMs {
		rate := float64(clip.SourceRange.DurationMs) / float64(clip.Range.DurationMs)
		filters = append(filters, timelineATempoFilters(rate)...)
	}
	filters = append(filters,
		"atrim=duration="+timelineSeconds(clip.Range.DurationMs),
		"volume="+timelineFloat(volume),
	)
	if clip.Range.StartMs > 0 {
		filters = append(filters, fmt.Sprintf("adelay=%d:all=1", clip.Range.StartMs))
	}
	return fmt.Sprintf("[%d:a:0]%s[%s]", input, strings.Join(filters, ","), label)
}

func timelineATempoFilters(rate float64) []string {
	filters := make([]string, 0, 4)
	for rate > 100 {
		filters = append(filters, "atempo=100")
		rate /= 100
	}
	for rate < 0.5 {
		filters = append(filters, "atempo=0.5")
		rate /= 0.5
	}
	if math.Abs(rate-1) > 0.0000005 {
		filters = append(filters, "atempo="+timelineFloat(rate))
	}
	return filters
}

func timelineRotatedPosition(transform TimelineTransform) (int32, int32) {
	if transform.RotationDegrees == 0 {
		return transform.X, transform.Y
	}
	radians := transform.RotationDegrees * math.Pi / 180
	rotatedWidth := math.Abs(float64(transform.Width)*math.Cos(radians)) + math.Abs(float64(transform.Height)*math.Sin(radians))
	rotatedHeight := math.Abs(float64(transform.Width)*math.Sin(radians)) + math.Abs(float64(transform.Height)*math.Cos(radians))
	x := transform.X - int32(math.Round((rotatedWidth-float64(transform.Width))/2))
	y := transform.Y - int32(math.Round((rotatedHeight-float64(transform.Height))/2))
	return x, y
}

func escapeTimelineFilterPath(path string) string {
	return strings.NewReplacer("\\", "\\\\", "'", "\\'", ":", "\\:", ",", "\\,").Replace(path)
}

func timelineSeconds(milliseconds int64) string {
	return timelineFloat(float64(milliseconds) / 1000)
}

func timelineFloat(value float64) string {
	if math.Abs(value) < 0.0000005 {
		value = 0
	}
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func writeTimelineCommand(writer io.Writer, bin string, args []string) {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, strconv.Quote(bin))
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	_, _ = fmt.Fprintln(writer, strings.Join(parts, " "))
}
