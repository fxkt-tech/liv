# FFcut Renderer

## 1. Scope / Trigger

Use `ffcut/renderer` when an application has already resolved local assets into a valid `ffcut.Project` and needs a real MP4 artifact. The package owns the supported FFmpeg execution subset; application repositories must not rebuild its filter graph or shell command.

The executable subset targets deterministic short-form voice-over videos: ordered hard cuts, original clip audio disabled, center-cover/contain/stretch fitting, one full-timeline voice track, an optional full-timeline BGM track, and ordered static/animated/video/text foreground layers.

## 2. Signatures

```go
func renderer.Render(
    ctx context.Context,
    project *ffcut.Project,
    absoluteOutputPath string,
    opts ...renderer.Option,
) error

func renderer.WithFFmpegBin(bin string) renderer.Option
func renderer.WithVideoCRF(crf int) renderer.Option
func renderer.WithDebug(writer io.Writer) renderer.Option
```

Errors are classified by these sentinels:

```go
var renderer.ErrUnsupportedProject error
var renderer.ErrInvalidOutput error
var renderer.ErrRenderFailed error
```

## 3. Contracts

- `Render` always calls `Project.Validate` before examining renderer support.
- `absoluteOutputPath` must be an absolute path. The caller owns parent-directory creation and cleanup.
- Supported canvas background: `color`.
- Supported transitions: `cut` only.
- Supported video: at least one clip; `audio.enabled=false`; no loop or final-frame freeze; fit is `cover`, `contain`, or `stretch`.
- Supported audio: exactly one `AudioTrackKindVoice` and at most one `AudioTrackKindBGM`. Both span the complete timeline from zero without loop or fades. Track slice order has no semantic meaning; the renderer resolves each input by `kind`.
- With BGM, the renderer applies each track's gain and mixes with `amix=inputs=2:duration=first:dropout_transition=0:normalize=0`. Voice remains the duration owner and is not implicitly normalized down. Without BGM, the historical single-voice graph is unchanged.
- Supported foreground media: static image, looping animation, and looping video. Every source starts at frame zero when its absolute layer range begins; media source audio is ignored.
- Supported foreground text: subtitle cues rendered in layer order with pixel/percent geometry, fill/background, alignment, and optional stroke. Cue text must be passed through temporary text files, never interpolated directly into an FFmpeg filter graph.
- Layer geometry uses the Project canvas coordinate system. Opacity and rotation are applied before overlay, and the layer is enabled only inside its absolute range.
- Output: H.264 High Profile, `yuv420p`, AAC 48 kHz, MP4 fast-start; CRF defaults to 20 and audio bitrate to 192k.
- Rendering is silent by default. `WithDebug` writes the quoted command to the supplied writer; it must not enable implicit stdout/stderr logging.
- Cancellation is controlled exclusively by `context.Context` and returns an error matching `ErrRenderFailed` while preserving `ctx.Err()` in the chain.

## 4. Validation & Error Matrix

| Condition | Required result |
|---|---|
| Invalid Project v2 | Return the original ffcut validation category before running FFmpeg |
| Empty or relative output path | Error matching `ErrInvalidOutput` |
| Image/blur background, fade transition, enabled clip audio, clip loop/freeze, BGM-only, multiple voice tracks, or multiple BGM tracks | Error matching `ErrUnsupportedProject` with the field path |
| Unsupported foreground media kind or malformed media/text payload | Protocol validation error, or `ErrUnsupportedProject` if valid but outside the renderer subset |
| Voice does not cover the exact video timeline | Error matching `ErrUnsupportedProject` |
| BGM does not cover the exact video timeline, loops, or fades | Error matching `ErrUnsupportedProject` |
| FFmpeg exits non-zero | Error matching `ErrRenderFailed`, wrapping the execution error and bounded stderr returned by the process |
| Context cancelled or deadline exceeded | Error matching both `ErrRenderFailed` and the context error |

## 5. Good / Base / Bad Cases

- Good: two portrait or landscape source clips center-cropped into a 9:16 canvas, hard-cut at the exact boundary, with one narration track and an optional lower-gain BGM track.
- Good: `examples/ffcut/renderer` builds this two-clip voice-over Project from local files, validates it with `ffcut.Marshal`, and renders a real MP4.
- Good: one clip with `FitModeContain` and a color canvas.
- Good: static PNG, looping GIF, looping silent video, and stroked text composited in slice order over the voice-over timeline.
- Base: one rate-1 clip and one voice track with identical two-second ranges.
- Bad: preserving source audio and mixing narration in the application; clip audio is intentionally unsupported in this subset.
- Bad: calling `exec.Command("ffmpeg", ...)` in an application repository for a Project that this renderer owns.
- Bad: passing an OSS URL or relative path as a `LocalSource`; compilation must resolve and download assets first.

## 6. Tests Required

- Unit-test command construction for one and multiple clips, voice-only and voice-plus-BGM in either slice order, every supported fit, custom CRF, command failure, invalid output, cancellation, and every unsupported feature.
- Integration-test with generated MP3/WAV narration and M4A BGM sources. Assert output dimensions, approximate duration, exactly one video and one mixed audio stream, and frame colors on both sides of the cut.
- Integration-test generated static image, multi-frame GIF, and video layers. Assert layer order, animation looping, video looping, and safe punctuation in text.
- Assert original clip audio is absent from the rendered output contract.
- Run `go test -race ./ffcut ./ffcut/renderer`, `go vet ./ffcut ./ffcut/renderer`, and the renderer integration test on a host with FFmpeg/FFprobe.

## 7. Wrong vs Correct

### Wrong

```go
args := []string{"-i", clip, "-filter_complex", applicationGraph, output}
err := exec.CommandContext(ctx, "ffmpeg", args...).Run()
```

This duplicates renderer policy, bypasses Project validation, and makes codec/output behavior application-specific.

### Correct

```go
if _, err := ffcut.Marshal(project); err != nil {
    return err
}
if err := renderer.Render(ctx, project, absoluteOutputPath); err != nil {
    return fmt.Errorf("render ffcut project: %w", err)
}
```

The application constructs immutable business input; ffcut validates the protocol; the renderer exclusively owns executable FFmpeg behavior.
