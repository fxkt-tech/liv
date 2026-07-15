package ffvmix

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fxkt-tech/liv/ffcut"
	"github.com/fxkt-tech/liv/ffprobe"
)

type mediaStreamMetadata struct {
	Width    int32
	Height   int32
	Duration time.Duration
}

type probeMetadata struct {
	Video          *mediaStreamMetadata
	Audio          *mediaStreamMetadata
	FormatDuration time.Duration
}

type mediaProber interface {
	Probe(context.Context, string) (probeMetadata, error)
}

type localFFprober struct{}

func (localFFprober) Probe(ctx context.Context, path string) (probeMetadata, error) {
	result, err := ffprobe.New().Input(path).Extract(ctx)
	if err != nil {
		return probeMetadata{}, err
	}
	metadata := probeMetadata{}
	if stream := result.GetFirstVideoStream(); stream != nil {
		metadata.Video = &mediaStreamMetadata{
			Width:    stream.Width,
			Height:   stream.Height,
			Duration: stream.Duration.Std(),
		}
	}
	if stream := result.GetFirstAudioStream(); stream != nil {
		metadata.Audio = &mediaStreamMetadata{Duration: stream.Duration.Std()}
	}
	if format := result.GetFormat(); format != nil {
		metadata.FormatDuration = format.Duration.Std()
	}
	return metadata, nil
}

type CompileOption func(*compileOptions)

type compileOptions struct {
	baseDir           string
	strictFingerprint bool
	probeConcurrency  int
	prober            mediaProber
}

func defaultCompileOptions() compileOptions {
	concurrency := min(runtime.GOMAXPROCS(0), 4)
	if concurrency < 1 {
		concurrency = 1
	}
	return compileOptions{
		probeConcurrency: concurrency,
		prober:           localFFprober{},
	}
}

func WithBaseDir(directory string) CompileOption {
	return func(options *compileOptions) {
		options.baseDir = directory
	}
}

func WithStrictFingerprint(enabled bool) CompileOption {
	return func(options *compileOptions) {
		options.strictFingerprint = enabled
	}
}

func WithProbeConcurrency(maximum int) CompileOption {
	return func(options *compileOptions) {
		options.probeConcurrency = maximum
	}
}

func withMediaProber(prober mediaProber) CompileOption {
	return func(options *compileOptions) {
		options.prober = prober
	}
}

type assetReference struct {
	fieldPath string
}

type assetRequest struct {
	path       string
	references []assetReference
	media      bool
}

type resolvedTemplateAssets struct {
	backgroundImage string
	videos          map[*VideoSource]string
	bgms            map[*BGM]string
	images          map[*LayerSpec]string
	subtitles       map[*LayerSpec]string
	fonts           map[*LayerSpec]string
	requests        map[string]*assetRequest
	order           []string
}

func newResolvedTemplateAssets() *resolvedTemplateAssets {
	return &resolvedTemplateAssets{
		videos:    make(map[*VideoSource]string),
		bgms:      make(map[*BGM]string),
		images:    make(map[*LayerSpec]string),
		subtitles: make(map[*LayerSpec]string),
		fonts:     make(map[*LayerSpec]string),
		requests:  make(map[string]*assetRequest),
	}
}

func collectTemplateAssets(template *Template, options compileOptions, issues *[]Issue) *resolvedTemplateAssets {
	resolved := newResolvedTemplateAssets()
	if template == nil {
		return resolved
	}
	if options.baseDir != "" && !isAbsoluteLocalPath(options.baseDir) {
		*issues = append(*issues, Issue{
			Code:    IssuePathResolution,
			Path:    "compile.base_dir",
			Message: "must be an absolute local path",
			Cause:   ErrInvalidTemplate,
		})
	}

	add := func(fieldPath, rawPath string, media bool) string {
		path, err := resolveTemplatePath(options.baseDir, rawPath)
		if err != nil {
			*issues = append(*issues, Issue{
				Code:      IssuePathResolution,
				Path:      fieldPath,
				LocalPath: rawPath,
				Message:   "cannot resolve local path",
				Cause:     err,
			})
			return ""
		}
		request, exists := resolved.requests[path]
		if !exists {
			request = &assetRequest{path: path}
			resolved.requests[path] = request
			resolved.order = append(resolved.order, path)
		}
		request.references = append(request.references, assetReference{fieldPath: fieldPath})
		request.media = request.media || media
		return path
	}

	if template.Background.Image != nil && strings.TrimSpace(template.Background.Image.Path) != "" {
		resolved.backgroundImage = add("background.image.path", template.Background.Image.Path, true)
	}
	for slotIndex, slot := range template.Slots {
		if slot == nil {
			continue
		}
		for videoIndex, video := range slot.Videos {
			if video == nil || strings.TrimSpace(video.Path) == "" {
				continue
			}
			fieldPath := fmt.Sprintf("slots[%d].videos[%d].path", slotIndex, videoIndex)
			resolved.videos[video] = add(fieldPath, video.Path, true)
		}
	}
	for index, bgm := range template.BGMs {
		if bgm == nil || strings.TrimSpace(bgm.Path) == "" {
			continue
		}
		resolved.bgms[bgm] = add(fmt.Sprintf("bgms[%d].path", index), bgm.Path, true)
	}
	for index, layer := range template.Layers {
		if layer == nil {
			continue
		}
		if layer.Image != nil && strings.TrimSpace(layer.Image.Path) != "" {
			resolved.images[layer] = add(fmt.Sprintf("layers[%d].image.path", index), layer.Image.Path, true)
		}
		if layer.Subtitle == nil {
			continue
		}
		if layer.Subtitle.Input.SRT != nil && strings.TrimSpace(layer.Subtitle.Input.SRT.Path) != "" {
			resolved.subtitles[layer] = add(fmt.Sprintf("layers[%d].subtitle.input.srt.path", index), layer.Subtitle.Input.SRT.Path, false)
		}
		if layer.Subtitle.Input.ASS != nil && strings.TrimSpace(layer.Subtitle.Input.ASS.Path) != "" {
			resolved.subtitles[layer] = add(fmt.Sprintf("layers[%d].subtitle.input.ass.path", index), layer.Subtitle.Input.ASS.Path, false)
		}
		if strings.TrimSpace(layer.Subtitle.Style.FontPath) != "" {
			resolved.fonts[layer] = add(fmt.Sprintf("layers[%d].subtitle.style.font_path", index), layer.Subtitle.Style.FontPath, false)
		}
	}
	return resolved
}

func resolveTemplatePath(baseDirectory, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is empty")
	}
	if isAbsoluteLocalPath(path) {
		return cleanLocalPath(path), nil
	}
	if hasURIScheme(path) {
		return "", fmt.Errorf("remote paths are not supported")
	}
	if baseDirectory == "" {
		return "", fmt.Errorf("relative path requires WithBaseDir")
	}
	if !isAbsoluteLocalPath(baseDirectory) {
		return "", fmt.Errorf("base directory must be absolute")
	}
	return cleanLocalPath(joinLocalPath(baseDirectory, path)), nil
}

type assetInspection struct {
	asset          CompiledAsset
	metadata       probeMetadata
	statErr        error
	fingerprintErr error
	probeErr       error
}

func inspectAssets(ctx context.Context, resolved *resolvedTemplateAssets, options compileOptions, issues *[]Issue) map[string]assetInspection {
	inspections := make(map[string]assetInspection, len(resolved.order))
	if len(resolved.order) == 0 {
		return inspections
	}
	if options.probeConcurrency <= 0 {
		*issues = append(*issues, Issue{
			Code:    IssueInvalidValue,
			Path:    "compile.probe_concurrency",
			Message: "must be positive",
			Cause:   ErrInvalidTemplate,
		})
		return inspections
	}
	if options.prober == nil {
		*issues = append(*issues, Issue{
			Code:    IssueInvalidValue,
			Path:    "compile.prober",
			Message: "is required",
			Cause:   ErrInvalidTemplate,
		})
		return inspections
	}

	type result struct {
		path       string
		inspection assetInspection
	}
	jobs := make(chan *assetRequest)
	results := make(chan result, len(resolved.order))
	workerCount := min(options.probeConcurrency, len(resolved.order))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for request := range jobs {
				results <- result{path: request.path, inspection: inspectAsset(ctx, request, options)}
			}
		}()
	}
	go func() {
		for _, path := range resolved.order {
			jobs <- resolved.requests[path]
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()
	for result := range results {
		inspections[result.path] = result.inspection
	}

	for _, path := range resolved.order {
		request := resolved.requests[path]
		inspection := inspections[path]
		fieldPath := request.references[0].fieldPath
		switch {
		case inspection.statErr != nil:
			code := IssueFileStat
			message := "cannot inspect file"
			if errors.Is(inspection.statErr, context.Canceled) || errors.Is(inspection.statErr, context.DeadlineExceeded) {
				code = IssueCanceled
				message = "asset inspection canceled"
			}
			*issues = append(*issues, Issue{Code: code, Path: fieldPath, LocalPath: path, Message: message, Cause: inspection.statErr})
		case inspection.fingerprintErr != nil:
			*issues = append(*issues, Issue{Code: IssueFingerprint, Path: fieldPath, LocalPath: path, Message: "cannot fingerprint file", Cause: inspection.fingerprintErr})
		case inspection.probeErr != nil:
			*issues = append(*issues, Issue{Code: IssueProbe, Path: fieldPath, LocalPath: path, Message: "ffprobe failed", Cause: inspection.probeErr})
		}
	}
	return inspections
}

func inspectAsset(ctx context.Context, request *assetRequest, options compileOptions) assetInspection {
	inspection := assetInspection{}
	if err := ctx.Err(); err != nil {
		inspection.statErr = err
		return inspection
	}
	info, err := os.Stat(request.path)
	if err != nil {
		inspection.statErr = err
		return inspection
	}
	if !info.Mode().IsRegular() {
		inspection.statErr = fmt.Errorf("not a regular file")
		return inspection
	}
	if info.Size() <= 0 {
		inspection.statErr = fmt.Errorf("file is empty")
		return inspection
	}
	if info.ModTime().UnixNano() == 0 {
		inspection.statErr = fmt.Errorf("file modification time must not be zero")
		return inspection
	}
	inspection.asset = CompiledAsset{
		Path: request.path,
		Fingerprint: ffcut.MediaFingerprint{
			Size:             info.Size(),
			ModifiedUnixNano: info.ModTime().UnixNano(),
		},
	}
	if options.strictFingerprint {
		sha, err := fileSHA256(request.path)
		if err != nil {
			inspection.fingerprintErr = err
			return inspection
		}
		inspection.asset.Fingerprint.SHA256 = sha
	}
	if request.media {
		metadata, err := options.prober.Probe(ctx, request.path)
		if err != nil {
			inspection.probeErr = err
			return inspection
		}
		inspection.metadata = metadata
	}
	return inspection
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func assetFingerprint(asset CompiledAsset) string {
	values := []string{
		asset.Path,
		fmt.Sprintf("%d", asset.Fingerprint.Size),
		fmt.Sprintf("%d", asset.Fingerprint.ModifiedUnixNano),
		asset.Fingerprint.SHA256,
	}
	hash := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(hash[:])
}
