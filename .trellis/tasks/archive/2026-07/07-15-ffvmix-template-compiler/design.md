# FFVMix Template Compiler Design

## Public Flow

```go
template := ffvmix.NewTemplate(config)
slot, err := template.AddSlot(slotConfig)
candidate, err := slot.AddVideo(sourceConfig)

compiled, err := ffvmix.Compile(ctx, template,
    ffvmix.WithBaseDir(dir),
    ffvmix.WithStrictFingerprint(false),
)
```

Constructors allocate IDs once. Direct JSON decoding is allowed, but `Compile` rejects missing IDs and never normalizes by mutating the caller's template.

## Compile Stages

1. Structural validation with stable field paths.
2. Resolve local paths against explicit base directory.
3. Build a unique-path probe worklist.
4. Probe video, BGM, background-image and image-layer assets with bounded concurrency and collect all failures; fonts and subtitle text files are fingerprinted without FFprobe.
5. Validate first video stream and optional first audio stream.
6. Normalize duration and source ranges to `time.Duration`.
7. Capture fast or strict file fingerprints.
8. Parse subtitle files and validate global layer assets.
9. Compile each slot/source duration policy into an immutable adaptation plan.
10. Validate adjacent joins and precompute transition compatibility.
11. Return a new `CompiledTemplate`; share no mutable slices or maps with Template.

## Probe Seam

Production uses the repository `ffprobe` package directly. A minimal internal prober interface exists only at the I/O seam so unit tests can provide deterministic metadata without executing a process. It does not become part of the public Template interface.

`ffprobe` must not import the root `ffmpeg` package. That seemingly small reuse creates a transitive dependency from `ffvmix` to renderer/filter code.

The ffprobe model must stop using `float32` duration. Prefer parsing the original decimal duration into `time.Duration` through a checked decimal/string path; `float64` is acceptable only as an intermediate compatibility step.

## Duration Plans

Each compiled source stores one of:

- natural duration;
- exact speed-up/down with checked rate;
- deterministic trim parameters, with random offset deferred as a seed-derived pure calculation;
- loop plan;
- freeze-last-frame plan;
- infeasible reason.

Infeasible sources remain represented for diagnostics but never become generator choices. A required slot with no feasible source is a compile error.

## Global Timing Feasibility

Transition compatibility is a graph over adjacent video candidates. The compiler uses a shortest-path dynamic program over that graph to find the minimum valid output duration without materializing combinations. Absolute layers, slot-local layers, subtitle cues and BGM start/fade settings must fit their minimum applicable duration so a fixed global configuration cannot silently invalidate only shorter generated results.

## Immutable Result

`CompiledTemplate` exposes read-only query methods needed by the generator. Concrete collections stay unexported. Returned views are copies or immutable values; no method permits adding slots or candidates.

## Error Model

`CompileError` aggregates `Issue` values containing code, template field path, optional local path and wrapped cause. File-system, probe, parse and semantic errors retain their original causes for `errors.Is/As`.

## File Fingerprints

Fast mode records normalized path, size and nanosecond mtime. Strict mode additionally streams SHA-256 once per unique file. Original relative paths remain in Template; resolved absolute paths exist only in CompiledTemplate and generated local-source Project entries.
