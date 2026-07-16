# Directory Structure

> Package ownership and file placement for this Go media-processing library.

## Project Shape

Liv is a single Go module (`github.com/fxkt-tech/liv`), not an HTTP service. It
has no route, controller, repository, or database layers. Organize changes by
the media-domain package that owns the behavior.

```text
./                         High-level transcode and snapshot facade (`package liv`)
├── ffmpeg/                FFmpeg command construction and execution
│   ├── codec/             Codec and container constants
│   ├── filter/            Filter graph primitives
│   │   └── fsugar/        Higher-level filter combinations
│   ├── input/             Input argument construction
│   ├── output/            Output argument construction
│   ├── stream/            Stream labels and mappings
│   └── util/              FFmpeg-specific helpers
├── ffprobe/               FFprobe execution and typed media metadata
├── ffcut/                 Renderer-independent FFcut project v2 protocol
│   └── fusion/            Separate track-data composition API
├── ffvmix/                Template compilation and lazy project generation
│   └── constraints/       Pure constraint contracts and built-ins
├── fftool/                Small media tools such as masks and names
├── pkg/                   Domain-neutral reusable helpers
├── examples/              Runnable examples grouped by public package/use case
└── docs/                  Project notes and command examples
```

Reference boundaries:

- `ffcut/project.go`, `timeline.go`, and `layer.go` own the persisted FFcut v2
  shape; `validate.go` and `marshal.go` own its trust boundary.
- `ffvmix/template.go` owns editable template data, `compile.go` and `probe.go`
  own local-file compilation, and `generator.go` plus `project.go` own lazy
  enumeration and FFcut project construction.
- `ffmpeg/ffmpeg.go` orchestrates the argument objects owned by `input/`,
  `filter/`, and `output/`.
- Root files such as `transcode.go` and `snapshot.go` compose the lower-level
  packages into convenience services.

## Placement Rules

1. Put a new public concept in the package that owns its invariant. For
   example, an FFcut wire field belongs in `ffcut`; a template-only selection
   policy belongs in `ffvmix`.
2. Keep validation and codecs beside the type they protect. The established
   examples are `ffcut/validate.go` + `ffcut/marshal.go` and
   `ffvmix/validate.go` + `ffvmix/codec.go`.
3. Keep external process execution at adapter boundaries. FFprobe execution is
   in `ffprobe/ffprobe.go`; FFVMix depends on it through the private
   `mediaProber` interface in `ffvmix/probe.go`.
4. Put helpers under `pkg/` only when they are domain-neutral and reused.
   FFmpeg expression or stream helpers stay under `ffmpeg/`; FFVMix scheduling
   helpers stay under `ffvmix/`.
5. Group runnable usage under `examples/` by public package and feature, as in
   `examples/ffmpeg/filter/scale/main.go` and
   `examples/service/transcode/simplemp4/main.go`. Tests remain next to the
   code as `*_test.go`.
6. Do not add FFcut v2 protocol fields to `ffcut/fusion`; it is a distinct
   track-data API. FFVMix output targets `ffcut.Project`.

The current dependency direction is intentional:

```text
root liv facade -> ffmpeg + ffprobe
ffvmix         -> ffcut + ffprobe + ffvmix/constraints
ffcut          -> standard library only
ffmpeg         -> ffmpeg subpackages + small shared helpers
```

Avoid reversing these edges. In particular, `ffcut` is a renderer-independent
protocol and must not import FFmpeg execution code.

## File and API Conventions

- Package and directory names are short lower-case domain names: `ffcut`,
  `ffprobe`, `constraints`.
- File names are lower-case topic names, with underscores for compound topics:
  `track_item.go`, `quality-guidelines.md`, `generator_test.go`.
- Constructors and configurable public operations use package-local functional
  options where already established:

  ```go
  ffmpeg.New(ffmpeg.WithBin("ffmpeg"), ffmpeg.WithDebug(true))
  ffvmix.Compile(ctx, template, ffvmix.WithBaseDir(baseDir))
  ffvmix.NewGenerator(compiled, ffvmix.WithSeed(42))
  ```

- Put package-wide sentinels and structured error types in `errors.go`.
- Prefer one file per coherent concept, not one file per exported type. The
  `ffvmix` package is the reference for a larger module split.

## Pre-Change Questions

- Which package owns the invariant being changed?
- Is this persisted protocol, editable configuration, compiled state, or
  execution behavior?
- Would the change introduce a reverse dependency or external I/O into a pure
  package?
- Do the co-located tests and package-specific Trellis spec need to change?
