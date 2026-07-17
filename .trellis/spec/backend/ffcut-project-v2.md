# FFcut Project v2 Protocol

## 1. Scope / Trigger

Use this contract whenever code constructs, serializes, validates, or consumes `ffcut.Project`. The root `ffcut` package is a pure, renderer-independent boundary: it performs no filesystem access, probing, or FFmpeg work.

This matters because `ffvmix` must emit an execution-ready timeline. Ambiguous timing or unresolved paths must fail before a renderer sees the Project.

## 2. Signatures

```go
const ProjectVersion = 2

func NewDuration(value time.Duration) (ffcut.Duration, error)
func DurationFromMicroseconds(value int64) (ffcut.Duration, error)

func (p *ffcut.Project) Validate() error
func ffcut.Marshal(project *ffcut.Project) ([]byte, error)
func ffcut.Unmarshal(data []byte) (*ffcut.Project, error)
```

Call `ffcut.Marshal`, not `encoding/json.Marshal`, at persistence or process boundaries. `ffcut.Marshal` validates the complete aggregate first; `ffcut.Unmarshal` rejects unknown fields and validates after decoding.

## 3. Contracts

### Aggregate shape

| Value | Required content |
|-------|------------------|
| `Project` | Version, ID, canvas, primary video sequence, metadata; optional BGM tracks and ordered layers |
| `Canvas` | Positive width/height, positive rational FPS, one typed background |
| `VideoClip` | Local source, source/timeline ranges, rate/loop/freeze behavior, fit mode, original-audio settings |
| `Transition` | Adjacent clip IDs, absolute range, `cut` or `fade`, audio-crossfade decision |
| `AudioTrack` | `bgm` or `voice` source/ranges, loop, gain, fade-in and fade-out |
| `Layer` | Absolute range plus exactly one image, media, or subtitle payload; slice order is render order |
| `Metadata` | Template fingerprint, actual seed, combination fingerprint, selections and constraint fingerprints |

### Time

- Go values use the named `ffcut.Duration`, backed by `time.Duration`.
- JSON values are integer microseconds. Sub-microsecond Go values, negative durations, and overflow are invalid.
- `TimeRange` is absolute `start + duration`; starts are non-negative and normal content durations are positive.
- For a non-looping video clip:

  ```text
  source_range.duration / playback.rate
      + playback.freeze_last_frame
      == timeline_range.duration
  ```

  Equality allows at most one microsecond of floating-point conversion tolerance. Loop and freeze are mutually exclusive.

### Sources

- `LocalSource.Path` is an absolute OS-local path. Template-relative paths must be resolved by `ffvmix.Compile` before Project construction.
- Validation checks path and fingerprint shape only; it never stats or opens the file.
- Fingerprints require positive file size and non-zero modification time. Optional SHA-256 is exactly 32 bytes encoded as 64 hexadecimal characters.

### Typed unions

`Background` and `Layer` use a discriminator plus typed pointer payloads. Exactly one payload must be non-nil and it must match the discriminator. Do not introduce `map[string]any`, unrestricted `interface{}`, or renderer-specific filter payloads.

`MediaLayer` is the reusable foreground-media payload. It carries a local source, pixel/percent geometry, opacity, rotation, and an explicit media kind:

- `image` is a static image and must not loop.
- `animation` and `video` must set `loop=true`; each layer range starts the source at frame zero.
- Layer media audio is never part of the protocol mix. Audio continues to come from `Project.Audio`.

`SubtitleStyle` supports fill/background color, alignment, font size, and optional stroke color/width. `SubtitleLayer` optionally carries opacity (omitted means fully opaque for existing Project v2 payloads) and rotation. Text content remains data, not an executable renderer expression.

## 4. Validation & Error Matrix

| Condition | Required result |
|-----------|-----------------|
| Nil Project, missing required ID/content, invalid enum or union | Error matching `ffcut.ErrInvalidProject` with a field path |
| Version other than 2 | Also matches `ffcut.ErrUnsupportedVersion` |
| Negative, overflowing, or sub-microsecond protocol time | Also matches `ffcut.ErrInvalidDuration` |
| Relative path or URL in `LocalSource.Path` | Invalid Project at `<source>.path` |
| Non-looping playback does not satisfy the duration equation | Invalid Project at `<clip>.playback` |
| Transition does not join the two adjacent clips exactly | Invalid Project at the transition field/range |
| Layer or BGM extends beyond the video sequence | Invalid Project at its timeline range |
| Unknown JSON field, malformed JSON, or trailing JSON value | `*ffcut.CodecError` matching `ffcut.ErrInvalidProject` |
| Multiple independent violations | One `*ffcut.ValidationError` containing all discovered field-path issues |

`AudioTrackKindVoice` and `SelectionKindVoice` are valid protocol values. Like video and BGM selections, a voice selection must carry a non-empty `asset_fingerprint`; the root protocol does not assume which renderer will consume it.

Errors must remain usable through `errors.Is`/`errors.As`; callers must not parse error strings.

## 5. Good / Base / Bad Cases

- Good: a 10-second source played at rate 2 into a 5-second timeline clip.
- Good: a 4-second source in a 5-second timeline clip with one second of final-frame freeze.
- Good: a short source with `loop=true` and no freeze.
- Base: a source and timeline range of equal duration at rate 1.
- Bad: a relative media path such as `media/a.mp4`; resolution belongs in template compilation.
- Bad: a 4-second source in a 5-second non-looping clip at rate 1 without freeze.
- Bad: a layer with both media and subtitle payloads, even if its discriminator names one of them.
- Bad: an image with `loop=true`, or an animation/video with `loop=false`.

## 6. Tests Required

- Round-trip a complete Project and assert semantic deep equality and layer order.
- Assert an exact minimal Project v2 JSON shape, including integer-microsecond fields.
- Assert serialization does not mutate the caller's Project.
- Table-test every validation category and assert both sentinel category and field path.
- Cover cut/fade adjacency, speed/loop/freeze playback, all background variants, BGM bounds, and subtitle/image/media layer bounds.
- Run `go test -race ./ffcut`, `go vet ./ffcut`, and `staticcheck ./ffcut` for protocol changes.
- Verify `ffcut` does not import `ffprobe`, `ffmpeg`, or `ffcut/fusion`.

## 7. Wrong vs Correct

### Wrong

```go
project.Video.Clips[0] = ffcut.VideoClip{
    Source:        ffcut.LocalSource{Path: "media/a.mp4"},
    SourceRange:   sourceRange(4 * time.Second),
    TimelineRange: timelineRange(5 * time.Second),
    Playback:      ffcut.Playback{Rate: 1},
}

data, _ := json.Marshal(project) // bypasses aggregate validation
```

### Correct

```go
project.Video.Clips[0] = ffcut.VideoClip{
    Source:        compiledAbsoluteLocalSource,
    SourceRange:   sourceRange(4 * time.Second),
    TimelineRange: timelineRange(5 * time.Second),
    Playback: ffcut.Playback{
        Rate:            1,
        FreezeLastFrame: duration(1 * time.Second),
    },
}

data, err := ffcut.Marshal(project)
```

The correct form makes the renderer's behavior explicit and keeps validation at the shared protocol boundary.
