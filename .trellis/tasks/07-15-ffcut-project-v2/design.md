# FFcut Project v2 Design

## Boundary

The root `ffcut` package becomes a pure timeline protocol. It accepts no runtime dependencies and performs no I/O. The empty legacy `FFcut` facade can be replaced; the separate `ffcut/fusion` package remains unchanged during this child task.

## Core Shape

```go
type Project struct {
    Version  int
    Canvas   Canvas
    Video    Sequence
    Audio    []AudioTrack
    Layers   []Layer
    Metadata Metadata
}
```

`Sequence` contains ordered clips and transitions that reference adjacent clip IDs. Each clip contains an execution-ready source range and timeline range. Template slot IDs may appear only as provenance, never as scheduling input.

## Typed Unions

Background and Layer use a discriminator plus optional typed payload fields. Validation requires exactly one payload matching the discriminator. This keeps standard JSON encoding practical while avoiding unbounded polymorphic values.

```go
type Layer struct {
    Kind     LayerKind
    Range    TimeRange
    Image    *ImageLayer
    Subtitle *SubtitleLayer
}
```

## Time and Space

- `Duration` is a named protocol type backed by Go `time.Duration`; JSON encodes it as integer microseconds.
- Conversion helpers accept and return `time.Duration` with overflow checks.
- `TimeRange` uses start + duration and rejects negative values.
- `Length` is `{Value float64, Unit px|percent}`; percentages are limited to 0–100.
- Canvas width, height and FPS must be positive.

## Validation

Validation walks the aggregate once and returns an aggregate error with stable field paths. It checks:

- unique non-empty IDs;
- source and timeline ranges;
- consistency between source duration, playback rate, loop/freeze behavior and timeline duration;
- ordered/non-overlapping sequence semantics except declared transitions;
- transition adjacency and duration;
- absolute local source paths and fingerprint shape without touching the filesystem;
- union discriminator/payload agreement;
- layer timing, units and required content;
- metadata presence needed for provenance.

## JSON Contract

`Marshal(Project)` validates first. `Unmarshal([]byte)` uses `encoding/json`, rejects unknown enum values, validates the result and never fills missing IDs. JSON uses snake_case field names and a top-level version.

No codec method mutates or sorts the caller's slices.

## Error Contract

Expose sentinel categories for invalid project and unsupported version, wrapped with field-path details. Callers can use `errors.Is` without parsing strings.

## Rollout

Only root-package tests are required here because legacy fusion execution tests are non-hermetic. Later children consume this package before any renderer exists, keeping rollback to deletion of the new pure types.
