package renderer

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fxkt-tech/liv/ffcut"
)

const (
	videoCodec  = "libx264"
	audioCodec  = "aac"
	pixelFormat = "yuv420p"
	voiceGain   = 1.0
	maxBGMGain  = 1.0
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

	if err := validateSupported(project); err != nil {
		return err
	}
	textFiles, clearTextFiles, err := materializeSubtitleTexts(project)
	if err != nil {
		return err
	}
	defer clearTextFiles()

	args, err := buildArgs(project, outputPath, config, textFiles)
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

func buildArgs(project *ffcut.Project, outputPath string, config config, textFiles map[subtitleCueKey]string) ([]string, error) {
	if err := validateSupported(project); err != nil {
		return nil, err
	}
	audio, err := resolveAudioTracks(project)
	if err != nil {
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
	layerInputs := make(map[int]int)
	nextInput := len(project.Video.Clips)
	for index, layer := range project.Layers {
		source, mediaKind, ok := visualLayerSource(layer)
		if !ok {
			continue
		}
		args = append(args, visualLayerInputArgs(source.Path, mediaKind, layer.Range.Duration)...)
		layerInputs[index] = nextInput
		nextInput++
	}

	voice := project.Audio[audio.voiceIndex]
	voiceInput := nextInput
	args = append(args,
		"-ss", seconds(voice.SourceRange.Start),
		"-t", seconds(voice.SourceRange.Duration),
		"-i", voice.Source.Path,
	)
	nextInput++
	bgmInput := -1
	if audio.bgmIndex >= 0 {
		bgm := project.Audio[audio.bgmIndex]
		bgmInput = nextInput
		args = append(args,
			"-ss", seconds(bgm.SourceRange.Start),
			"-t", seconds(bgm.SourceRange.Duration),
			"-i", bgm.Source.Path,
		)
	}

	graph := make([]string, 0, len(project.Video.Clips)+len(project.Layers)*2+2)
	videoLabels := make([]string, 0, len(project.Video.Clips))
	for index, clip := range project.Video.Clips {
		label := fmt.Sprintf("v%d", index)
		videoLabels = append(videoLabels, fmt.Sprintf("[%s]", label))
		graph = append(graph, videoFilter(index, label, project.Canvas, clip))
	}
	baseLabel := "vout"
	if len(project.Layers) > 0 {
		baseLabel = "base"
	}
	if len(videoLabels) == 1 {
		graph = append(graph, videoLabels[0]+"null["+baseLabel+"]")
	} else {
		graph = append(graph, strings.Join(videoLabels, "")+fmt.Sprintf("concat=n=%d:v=1:a=0[%s]", len(videoLabels), baseLabel))
	}
	currentLabel := baseLabel
	for index, layer := range project.Layers {
		nextLabel := fmt.Sprintf("layerout%d", index)
		if index == len(project.Layers)-1 {
			nextLabel = "vout"
		}
		switch layer.Kind {
		case ffcut.LayerKindImage, ffcut.LayerKindMedia:
			inputLabel := fmt.Sprintf("layer%d", index)
			graph = append(graph, visualLayerFilter(layerInputs[index], inputLabel, project.Canvas, layer))
			graph = append(graph, overlayFilter(currentLabel, inputLabel, nextLabel, project.Canvas, layer))
		case ffcut.LayerKindSubtitle:
			graph = append(graph, subtitleLayerFilters(currentLabel, nextLabel, project.Canvas, layer, index, textFiles)...)
		}
		currentLabel = nextLabel
	}

	voiceOutputLabel := "aout"
	if audio.bgmIndex >= 0 {
		voiceOutputLabel = "voiceout"
	}
	graph = append(graph, fmt.Sprintf(
		"[%d:a:0]atrim=duration=%s,asetpts=PTS-STARTPTS,volume=%s[%s]",
		voiceInput,
		seconds(voice.TimelineRange.Duration),
		formatFloat(voice.Gain),
		voiceOutputLabel,
	))
	if audio.bgmIndex >= 0 {
		bgm := project.Audio[audio.bgmIndex]
		graph = append(graph,
			fmt.Sprintf(
				"[%d:a:0]atrim=duration=%s,asetpts=PTS-STARTPTS,volume=%s[bgmout]",
				bgmInput,
				seconds(bgm.TimelineRange.Duration),
				formatFloat(bgm.Gain),
			),
			"[voiceout][bgmout]amix=inputs=2:duration=first:dropout_transition=0:normalize=0[aout]",
		)
	}

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
	for index, layer := range project.Layers {
		switch layer.Kind {
		case ffcut.LayerKindImage, ffcut.LayerKindMedia, ffcut.LayerKindSubtitle:
		default:
			return unsupported(fmt.Sprintf("layers[%d].kind", index), string(layer.Kind))
		}
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
	audio, err := resolveAudioTracks(project)
	if err != nil {
		return err
	}
	projectDuration, _ := project.Video.Clips[len(project.Video.Clips)-1].TimelineRange.End()
	voice := project.Audio[audio.voiceIndex]
	if voice.SourceRange.Start != 0 || voice.TimelineRange.Start != 0 || voice.TimelineRange.Duration != projectDuration ||
		voice.Gain != voiceGain || voice.Loop || voice.FadeIn != 0 || voice.FadeOut != 0 {
		return unsupported(fmt.Sprintf("audio[%d]", audio.voiceIndex), "voice must start from zero, cover the complete timeline at gain 1, and have no loop or fades")
	}
	if audio.bgmIndex >= 0 {
		bgm := project.Audio[audio.bgmIndex]
		if bgm.SourceRange.Start != 0 || bgm.TimelineRange.Start != 0 || bgm.TimelineRange.Duration != projectDuration ||
			bgm.Gain > maxBGMGain || bgm.Loop || bgm.FadeIn != 0 || bgm.FadeOut != 0 {
			return unsupported(fmt.Sprintf("audio[%d]", audio.bgmIndex), "BGM must start from zero, cover the complete timeline at gain 0..1, and have no loop or fades")
		}
	}
	return nil
}

type audioTrackIndexes struct {
	voiceIndex int
	bgmIndex   int
}

func resolveAudioTracks(project *ffcut.Project) (audioTrackIndexes, error) {
	indexes := audioTrackIndexes{voiceIndex: -1, bgmIndex: -1}
	for index, track := range project.Audio {
		switch track.Kind {
		case ffcut.AudioTrackKindVoice:
			if indexes.voiceIndex >= 0 {
				return audioTrackIndexes{}, unsupported("audio", "exactly one voice track is required")
			}
			indexes.voiceIndex = index
		case ffcut.AudioTrackKindBGM:
			if indexes.bgmIndex >= 0 {
				return audioTrackIndexes{}, unsupported("audio", "at most one BGM track is supported")
			}
			indexes.bgmIndex = index
		}
	}
	if indexes.voiceIndex < 0 || len(project.Audio) > 2 {
		return audioTrackIndexes{}, unsupported("audio", "exactly one voice track and at most one BGM track are required")
	}
	return indexes, nil
}

func visualLayerSource(layer ffcut.Layer) (ffcut.LocalSource, ffcut.MediaKind, bool) {
	if layer.Image != nil {
		return layer.Image.Source, ffcut.MediaKindImage, true
	}
	if layer.Media != nil {
		return layer.Media.Source, layer.Media.Kind, true
	}
	return ffcut.LocalSource{}, "", false
}

func visualLayerInputArgs(path string, kind ffcut.MediaKind, duration ffcut.Duration) []string {
	args := make([]string, 0, 8)
	switch kind {
	case ffcut.MediaKindImage:
		args = append(args, "-loop", "1")
	case ffcut.MediaKindAnimation:
		args = append(args, "-ignore_loop", "0", "-stream_loop", "-1")
	case ffcut.MediaKindVideo:
		args = append(args, "-stream_loop", "-1")
	}
	return append(args, "-t", seconds(duration), "-i", path)
}

func visualLayerFilter(input int, label string, canvas ffcut.Canvas, layer ffcut.Layer) string {
	geometry, opacity, rotation := visualLayerProperties(layer)
	_, _, width, height := resolveGeometry(geometry, canvas)
	frameRate := fmt.Sprintf("%d/%d", canvas.FrameRate.Numerator, canvas.FrameRate.Denominator)
	filters := []string{
		fmt.Sprintf("scale=%d:%d", width, height),
		"setsar=1",
		"fps=" + frameRate,
		"format=rgba",
		fmt.Sprintf("colorchannelmixer=aa=%s", formatFloat(opacity)),
	}
	if rotation != 0 {
		angle := formatFloat(rotation) + "*PI/180"
		filters = append(filters, fmt.Sprintf("rotate=%s:c=none:ow=rotw(%s):oh=roth(%s)", angle, angle, angle))
	}
	filters = append(filters, "setpts=PTS-STARTPTS+"+seconds(layer.Range.Start)+"/TB")
	return fmt.Sprintf("[%d:v:0]%s[%s]", input, strings.Join(filters, ","), label)
}

func overlayFilter(currentLabel, inputLabel, outputLabel string, canvas ffcut.Canvas, layer ffcut.Layer) string {
	geometry, _, rotation := visualLayerProperties(layer)
	x, y := rotatedOverlayPosition(geometry, rotation, canvas)
	return overlayAt(currentLabel, inputLabel, outputLabel, x, y, layer.Range)
}

func overlayAt(currentLabel, inputLabel, outputLabel string, x, y int, value ffcut.TimeRange) string {
	return fmt.Sprintf(
		"[%s][%s]overlay=x=%d:y=%d:eof_action=pass:shortest=0:enable='%s'[%s]",
		currentLabel,
		inputLabel,
		x,
		y,
		timelineEnable(value),
		outputLabel,
	)
}

func visualLayerProperties(layer ffcut.Layer) (ffcut.Geometry, float64, float64) {
	if layer.Image != nil {
		return layer.Image.Geometry, layer.Image.Opacity, layer.Image.RotationDegrees
	}
	return layer.Media.Geometry, layer.Media.Opacity, layer.Media.RotationDegrees
}

func subtitleLayerFilters(currentLabel, outputLabel string, canvas ffcut.Canvas, layer ffcut.Layer, layerIndex int, textFiles map[subtitleCueKey]string) []string {
	subtitle := layer.Subtitle
	x, y, width, height := resolveGeometry(subtitle.Region, canvas)
	fontSize := resolveLength(subtitle.Style.FontSize, float64(canvas.Height))
	strokeWidth := resolveLength(subtitle.Style.StrokeWidth, float64(canvas.Height))
	frameRate := fmt.Sprintf("%d/%d", canvas.FrameRate.Numerator, canvas.FrameRate.Denominator)
	filters := []string{
		fmt.Sprintf("color=c=black@0.0:s=%dx%d:r=%s:d=%s", width, height, frameRate, seconds(layer.Range.Duration)),
		"format=rgba",
		"setpts=PTS-STARTPTS+" + seconds(layer.Range.Start) + "/TB",
	}
	for cueIndex, cue := range subtitle.Cues {
		xExpression := "0"
		switch subtitle.Style.Align {
		case ffcut.TextAlignCenter:
			xExpression = fmt.Sprintf("(%d-text_w)/2", width)
		case ffcut.TextAlignRight:
			xExpression = fmt.Sprintf("%d-text_w", width)
		}
		options := []string{
			"textfile='" + escapeDrawtextValue(textFiles[subtitleCueKey{layer: layerIndex, cue: cueIndex}]) + "'",
			"expansion=none",
			"x=" + xExpression,
			fmt.Sprintf("y=(%d-text_h)/2", height),
			"fontsize=" + formatFloat(fontSize),
			"fontcolor=" + subtitle.Style.Color,
			"enable='" + timelineEnable(cue.Range) + "'",
		}
		if subtitle.Style.Font != nil {
			options = append(options, "fontfile='"+escapeDrawtextValue(subtitle.Style.Font.Path)+"'")
		} else {
			options = append(options, "font='"+escapeDrawtextValue(subtitle.Style.FontFamily)+"'")
		}
		if subtitle.Style.BackgroundColor != "" {
			options = append(options, "box=1", "boxcolor="+subtitle.Style.BackgroundColor)
		}
		if strokeWidth > 0 {
			options = append(options, "borderw="+formatFloat(strokeWidth), "bordercolor="+subtitle.Style.StrokeColor)
		}
		filters = append(filters, "drawtext="+strings.Join(options, ":"))
	}
	opacity := 1.0
	if subtitle.Opacity != nil {
		opacity = *subtitle.Opacity
	}
	filters = append(filters, fmt.Sprintf("colorchannelmixer=aa=%s", formatFloat(opacity)))
	if subtitle.RotationDegrees != 0 {
		angle := formatFloat(subtitle.RotationDegrees) + "*PI/180"
		filters = append(filters, fmt.Sprintf("rotate=%s:c=none:ow=rotw(%s):oh=roth(%s)", angle, angle, angle))
	}
	inputLabel := fmt.Sprintf("subtitle%d", layerIndex)
	visual := fmt.Sprintf("%s[%s]", strings.Join(filters, ","), inputLabel)
	if subtitle.RotationDegrees != 0 {
		x, y = rotatedOverlayPosition(subtitle.Region, subtitle.RotationDegrees, canvas)
	}
	return []string{
		visual,
		overlayAt(currentLabel, inputLabel, outputLabel, x, y, layer.Range),
	}
}

type subtitleCueKey struct {
	layer int
	cue   int
}

func materializeSubtitleTexts(project *ffcut.Project) (map[subtitleCueKey]string, func(), error) {
	count := 0
	for _, layer := range project.Layers {
		if layer.Subtitle != nil {
			count += len(layer.Subtitle.Cues)
		}
	}
	if count == 0 {
		return map[subtitleCueKey]string{}, func() {}, nil
	}
	directory, err := os.MkdirTemp("", "ffcut-text-")
	if err != nil {
		return nil, func() {}, fmt.Errorf("%w: create subtitle text directory: %w", ErrInvalidOutput, err)
	}
	clear := func() { _ = os.RemoveAll(directory) }
	files := make(map[subtitleCueKey]string, count)
	for layerIndex, layer := range project.Layers {
		if layer.Subtitle == nil {
			continue
		}
		for cueIndex, cue := range layer.Subtitle.Cues {
			path := filepath.Join(directory, fmt.Sprintf("layer-%03d-cue-%03d.txt", layerIndex, cueIndex))
			if err = os.WriteFile(path, []byte(cue.Text), 0o600); err != nil {
				clear()
				return nil, func() {}, fmt.Errorf("%w: write subtitle text: %w", ErrInvalidOutput, err)
			}
			files[subtitleCueKey{layer: layerIndex, cue: cueIndex}] = path
		}
	}
	return files, clear, nil
}

func resolveGeometry(geometry ffcut.Geometry, canvas ffcut.Canvas) (int, int, int, int) {
	x := int(math.Round(resolveLength(geometry.X, float64(canvas.Width))))
	y := int(math.Round(resolveLength(geometry.Y, float64(canvas.Height))))
	width := int(math.Round(resolveLength(geometry.Width, float64(canvas.Width))))
	height := int(math.Round(resolveLength(geometry.Height, float64(canvas.Height))))
	switch geometry.Anchor {
	case ffcut.AnchorTopRight:
		x -= width
	case ffcut.AnchorBottomLeft:
		y -= height
	case ffcut.AnchorBottomRight:
		x -= width
		y -= height
	case ffcut.AnchorCenter:
		x -= width / 2
		y -= height / 2
	}
	return x, y, width, height
}

func rotatedOverlayPosition(geometry ffcut.Geometry, rotationDegrees float64, canvas ffcut.Canvas) (int, int) {
	x, y, width, height := resolveGeometry(geometry, canvas)
	if rotationDegrees == 0 {
		return x, y
	}
	radians := rotationDegrees * math.Pi / 180
	rotatedWidth := math.Abs(float64(width)*math.Cos(radians)) + math.Abs(float64(height)*math.Sin(radians))
	rotatedHeight := math.Abs(float64(width)*math.Sin(radians)) + math.Abs(float64(height)*math.Cos(radians))
	x -= int(math.Round((rotatedWidth - float64(width)) / 2))
	y -= int(math.Round((rotatedHeight - float64(height)) / 2))
	return x, y
}

func resolveLength(length ffcut.Length, axis float64) float64 {
	if length.Unit == ffcut.LengthUnitPercent {
		return length.Value * axis / 100
	}
	return length.Value
}

func timelineEnable(value ffcut.TimeRange) string {
	end, _ := value.End()
	return fmt.Sprintf("gte(t\\,%s)*lt(t\\,%s)", seconds(value.Start), seconds(end))
}

func escapeDrawtextValue(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		":", "\\:",
		",", "\\,",
		";", "\\;",
		"[", "\\[",
		"]", "\\]",
		"\n", "\\n",
	)
	return replacer.Replace(value)
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
