package renderer

import (
	"context"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fxkt-tech/liv/ffcut"
)

const (
	videoCodec  = "libx264"
	audioCodec  = "aac"
	pixelFormat = "yuv420p"
)

// Render validates project, builds the supported FFmpeg graph, and writes an MP4 file.
func Render(ctx context.Context, project *ffcut.Project, outputPath string, opts ...Option) error {
	if err := project.Validate(); err != nil {
		return err
	}
	if outputPath == "" || !filepath.IsAbs(outputPath) {
		return fmt.Errorf("%w: path must be absolute", ErrInvalidOutput)
	}

	config := defaultConfig()
	for _, option := range opts {
		if option != nil {
			option(&config)
		}
	}

	args, err := buildArgs(project, outputPath, config)
	if err != nil {
		return err
	}
	if config.debugWriter != nil {
		writeCommand(config.debugWriter, config.ffmpegBin, args)
	}

	output, err := config.runner.Run(ctx, config.ffmpegBin, args)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("%w: %w", ErrRenderFailed, ctx.Err())
	}
	stderr := strings.TrimSpace(string(output))
	if stderr == "" {
		return fmt.Errorf("%w: %w", ErrRenderFailed, err)
	}
	return fmt.Errorf("%w: %w: %s", ErrRenderFailed, err, stderr)
}

func buildArgs(project *ffcut.Project, outputPath string, config config) ([]string, error) {
	if err := validateSupported(project); err != nil {
		return nil, err
	}

	args := []string{"-v", "error", "-y"}
	for _, clip := range project.Video.Clips {
		args = append(args,
			"-ss", seconds(clip.SourceRange.Start),
			"-t", seconds(clip.SourceRange.Duration),
			"-i", clip.Source.Path,
		)
	}

	voice := project.Audio[0]
	args = append(args,
		"-ss", seconds(voice.SourceRange.Start),
		"-t", seconds(voice.SourceRange.Duration),
		"-i", voice.Source.Path,
	)

	graph := make([]string, 0, len(project.Video.Clips)+2)
	videoLabels := make([]string, 0, len(project.Video.Clips))
	for index, clip := range project.Video.Clips {
		label := fmt.Sprintf("v%d", index)
		videoLabels = append(videoLabels, fmt.Sprintf("[%s]", label))
		graph = append(graph, videoFilter(index, label, project.Canvas, clip))
	}
	if len(videoLabels) == 1 {
		graph = append(graph, videoLabels[0]+"null[vout]")
	} else {
		graph = append(graph, strings.Join(videoLabels, "")+fmt.Sprintf("concat=n=%d:v=1:a=0[vout]", len(videoLabels)))
	}

	voiceInput := len(project.Video.Clips)
	graph = append(graph, fmt.Sprintf(
		"[%d:a:0]atrim=duration=%s,asetpts=PTS-STARTPTS,volume=%s[aout]",
		voiceInput,
		seconds(voice.TimelineRange.Duration),
		formatFloat(voice.Gain),
	))

	frameRate := fmt.Sprintf("%d/%d", project.Canvas.FrameRate.Numerator, project.Canvas.FrameRate.Denominator)
	projectDuration, _ := project.Video.Clips[len(project.Video.Clips)-1].TimelineRange.End()
	args = append(args,
		"-filter_complex", strings.Join(graph, ";"),
		"-map", "[vout]",
		"-map", "[aout]",
		"-c:v", videoCodec,
		"-profile:v", "high",
		"-crf", strconv.Itoa(config.videoCRF),
		"-pix_fmt", pixelFormat,
		"-r", frameRate,
		"-c:a", audioCodec,
		"-b:a", config.audioBitrate,
		"-ar", strconv.Itoa(config.audioRate),
		"-movflags", "+faststart",
		"-t", seconds(projectDuration),
		outputPath,
	)
	return args, nil
}

func validateSupported(project *ffcut.Project) error {
	if project.Canvas.Background.Kind != ffcut.BackgroundKindColor {
		return unsupported("canvas.background.kind", string(project.Canvas.Background.Kind))
	}
	if len(project.Layers) != 0 {
		return unsupported("layers", "layers are not implemented")
	}
	for index, transition := range project.Video.Transitions {
		if transition.Kind != ffcut.TransitionKindCut {
			return unsupported(fmt.Sprintf("video.transitions[%d].kind", index), string(transition.Kind))
		}
	}
	for index, clip := range project.Video.Clips {
		if clip.Audio.Enabled {
			return unsupported(fmt.Sprintf("video.clips[%d].audio.enabled", index), "original clip audio")
		}
		if clip.Playback.Loop || clip.Playback.FreezeLastFrame != 0 {
			return unsupported(fmt.Sprintf("video.clips[%d].playback", index), "loop or freeze")
		}
	}
	if len(project.Audio) != 1 || project.Audio[0].Kind != ffcut.AudioTrackKindVoice {
		return unsupported("audio", "exactly one voice track is required")
	}
	voice := project.Audio[0]
	projectDuration, _ := project.Video.Clips[len(project.Video.Clips)-1].TimelineRange.End()
	if voice.TimelineRange.Start != 0 || voice.TimelineRange.Duration != projectDuration || voice.Loop || voice.FadeIn != 0 || voice.FadeOut != 0 {
		return unsupported("audio[0]", "voice must cover the complete timeline without loop or fades")
	}
	return nil
}

func videoFilter(index int, label string, canvas ffcut.Canvas, clip ffcut.VideoClip) string {
	width := strconv.Itoa(int(canvas.Width))
	height := strconv.Itoa(int(canvas.Height))
	filters := make([]string, 0, 6)
	switch clip.Fit {
	case ffcut.FitModeCover:
		filters = append(filters,
			fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=increase", width, height),
			fmt.Sprintf("crop=%s:%s:(iw-ow)/2:(ih-oh)/2", width, height),
		)
	case ffcut.FitModeContain:
		color := strings.TrimPrefix(canvas.Background.Color.Color, "#")
		filters = append(filters,
			fmt.Sprintf("scale=%s:%s:force_original_aspect_ratio=decrease", width, height),
			fmt.Sprintf("pad=%s:%s:(ow-iw)/2:(oh-ih)/2:color=0x%s", width, height, color),
		)
	case ffcut.FitModeStretch:
		filters = append(filters, fmt.Sprintf("scale=%s:%s", width, height))
	}
	frameRate := fmt.Sprintf("%d/%d", canvas.FrameRate.Numerator, canvas.FrameRate.Denominator)
	filters = append(filters,
		"setsar=1",
		"fps="+frameRate,
		fmt.Sprintf("setpts=(PTS-STARTPTS)/%s", formatFloat(clip.Playback.Rate)),
	)
	return fmt.Sprintf("[%d:v:0]%s[%s]", index, strings.Join(filters, ","), label)
}

func unsupported(path, feature string) error {
	return fmt.Errorf("%w: %s: %s", ErrUnsupportedProject, path, feature)
}

func seconds(duration ffcut.Duration) string {
	return formatFloat(duration.Std().Seconds())
}

func formatFloat(value float64) string {
	if math.Abs(value) < 0.0000005 {
		value = 0
	}
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func writeCommand(writer io.Writer, bin string, args []string) {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, strconv.Quote(bin))
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	_, _ = fmt.Fprintln(writer, strings.Join(parts, " "))
}
