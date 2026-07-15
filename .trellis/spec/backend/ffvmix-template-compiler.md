# FFVMix Template Compiler

## 1. Scope / Trigger

Use this contract when persisting an FFVMix template or converting one into a `CompiledTemplate`. Compilation is the only I/O boundary in FFVMix: it resolves paths, fingerprints files, runs FFprobe, parses subtitle files, and precomputes feasibility.

The generator must consume only `CompiledTemplate`. It must never resolve paths, stat files, run FFprobe, parse SRT/ASS, or reinterpret duration policies.

Dependency direction is strict:

```text
ffvmix -> ffcut
ffvmix -> ffprobe
ffprobe -X-> ffmpeg
ffvmix -X-> ffmpeg/filter
```

`ffprobe` owns its command-line log-level value. Importing the root `ffmpeg` package merely to reuse a log-level constant is forbidden because it pulls renderer/filter packages through the compiler boundary.

## 2. Signatures

```go
const ffvmix.TemplateSchemaVersion = 1

func ffvmix.NewTemplate(config ffvmix.TemplateConfig) *ffvmix.Template
func (t *ffvmix.Template) AddSlot(config ffvmix.SlotConfig) (*ffvmix.Slot, error)
func (s *ffvmix.Slot) AddVideo(config ffvmix.VideoSourceConfig) (*ffvmix.VideoSource, error)
func (t *ffvmix.Template) AddJoin(config ffvmix.JoinConfig) (*ffvmix.Join, error)
func (j *ffvmix.Join) AddTransition(config ffvmix.TransitionConfig) (*ffvmix.TransitionCandidate, error)
func (t *ffvmix.Template) AddBGM(config ffvmix.BGMConfig) (*ffvmix.BGM, error)

func ffvmix.MarshalTemplate(template *ffvmix.Template) ([]byte, error)
func ffvmix.UnmarshalTemplate(data []byte) (*ffvmix.Template, error)

func ffvmix.Compile(
    ctx context.Context,
    template *ffvmix.Template,
    options ...ffvmix.CompileOption,
) (*ffvmix.CompiledTemplate, error)

func ffvmix.WithBaseDir(absolutePath string) ffvmix.CompileOption
func ffvmix.WithStrictFingerprint(enabled bool) ffvmix.CompileOption
func ffvmix.WithProbeConcurrency(maximum int) ffvmix.CompileOption
```

`CompiledTemplate` exposes copy-returning getters for canvas, background, slots, joins, BGMs, ordered layers, and constraints. Its concrete collections remain unexported.

## 3. Contracts

### Persisted template

- Constructors generate UUID-backed IDs once. IDs are serialized and never synthesized while decoding or compiling.
- Direct `encoding/json` decoding is allowed for callers that need it, but missing IDs remain missing and `Compile` rejects them.
- `MarshalTemplate` and `UnmarshalTemplate` are validated, strict boundaries. Unknown fields and trailing JSON are errors.
- Template time fields reuse `ffcut.Duration` and serialize as integer microseconds.
- A slot is required and selects one video candidate. Joins are ordered and must connect each adjacent slot pair exactly once.
- Background, layer, subtitle input, layer timing, and built-in constraint specs are discriminator-plus-typed-payload unions. Exactly one matching payload is required.

### Path and asset boundary

- Relative template paths require an explicit absolute `WithBaseDir`; process CWD is never a fallback.
- URLs and URI-like schemes are rejected. Compiled paths are cleaned absolute local paths.
- Cleaned paths are de-duplicated before file inspection. Each unique video, audio, background image, or image layer is probed at most once per compile.
- Fonts and subtitle text files are statted/fingerprinted but are not sent to FFprobe.
- Fast fingerprints contain size plus nanosecond mtime. Strict mode adds one streamed SHA-256 per unique file.
- Video candidates require the first video stream. Their first audio stream is optional and absence means silence.
- BGM requires the first audio stream. Background and image layers require the first visual stream and positive dimensions.

### Probe time

`ffprobe.Duration` parses the original JSON decimal-second token through checked rational arithmetic into `time.Duration`. It does not pass through `float32` or `float64`. Positive media durations are floored to microseconds when entering the FFcut protocol so a source range never claims time beyond the real media boundary.

### Duration adaptation

For selected source duration `S` and optional target `D`:

| Condition | Policy | Compiled result |
|-----------|--------|-----------------|
| No target or `S == D` | — | Natural, rate 1 |
| `S > D` | `speed_up` | Rate `S/D` if at or below maximum |
| `S > D` | `trim` | Source duration `D`; start/center/end offset fixed, random offset deferred with a maximum |
| `S > D` | `reject` | Infeasible candidate with reason |
| `S < D` | `slow_down` | Rate `S/D` if at or above minimum |
| `S < D` | `loop` | Loop to timeline duration `D` |
| `S < D` | `freeze` | Rate 1 plus `D-S` final-frame freeze |
| `S < D` | `reject` | Infeasible candidate with reason |

Infeasible candidates remain visible in compiled slot views. A required slot with no feasible candidate fails compilation.

### Global timing

- Transition compatibility is precomputed for every transition and adjacent feasible video pair. Oversized transitions remain represented but incompatible.
- The compiler computes the shortest valid output with dynamic programming over that compatibility graph; it does not enumerate the Cartesian product.
- Absolute layers, full-output subtitle cues, slot-anchored layers, and BGM start/fade timing must fit the shortest applicable output or slot duration. A global configuration may not silently invalidate only some generated combinations.
- SRT and ASS files become the same ordered `[]NormalizedCue` representation. Cue IDs derived from file input are stable from layer ID plus cue order.

### Immutability

`Compile` never mutates the input Template. `CompiledTemplate` owns all derived data; getters deep-copy slices, maps, pointer payloads, subtitle cues, and constraint payloads. Querying a compiled result performs no I/O.

## 4. Validation & Error Matrix

| Condition | Required result |
|-----------|-----------------|
| Missing/duplicate ID, invalid enum, union, weight, rate, reference, or geometry | `*ffvmix.CompileError` matching `ErrInvalidTemplate`, with stable field path |
| Relative asset path without absolute base dir, or URI path | `IssuePathResolution` at the template path field |
| Missing/empty/non-regular file | `IssueFileStat`, preserving `os` cause for `errors.Is` |
| SHA read failure | `IssueFingerprint`, preserving the I/O cause |
| FFprobe command failure | `IssueProbe`, preserving wrapped process/context cause and stderr |
| Video without visual stream | `IssueMissingVideo` |
| BGM without audio stream | `IssueMissingAudio` |
| Configured source range outside probed duration | `IssueSourceRange` |
| Invalid SRT/ASS | `IssueSubtitleParse` with layer path and local path |
| No feasible video in a required slot | `IssueNoFeasibleSource` |
| No compatible transition across an adjacent pair | `IssueNoFeasibleTransition` |
| Canceled/deadline context | `IssueCanceled`; aggregate error matches the context error |
| Multiple independent failures | One deterministic `CompileError.Issues` list; do not stop at the first asset |

Callers use `errors.Is`/`errors.As`; they must not parse formatted error strings.

## 5. Good / Base / Bad Cases

- Good: two candidates reference `media/a.mp4`; the cleaned path is probed once and both candidates receive independent IDs/plans.
- Good: a video-only file compiles with `HasAudio=false`.
- Good: a 10-second source in a 5-second trim slot stores a 5-second source duration and a bounded trim offset.
- Good: an 8-second fade remains in the transition pool but is incompatible with a 5-second slot, while a 1-second fade remains usable.
- Base: a one-slot template with an absolute local video path, black background, no joins, BGM, or layers.
- Bad: resolving `media/a.mp4` from process CWD because no base directory was supplied.
- Bad: allowing a late absolute watermark or BGM that exists only in longer output combinations.
- Bad: returning internal slices/maps or retaining Template payload pointers in `CompiledTemplate`.

## 6. Tests Required

- Round-trip a complete Template and assert IDs, slice order, policies, and deep equality.
- Assert missing IDs, unknown JSON fields, invalid references, and union mismatches fail with field paths.
- Use a thread-safe fake prober to assert one call per unique cleaned media path and no calls after compilation.
- Aggregate file-missing, missing-stream, out-of-range, and invalid-reference failures in one compile test.
- Cover natural, speed-up, slow-down, trim, loop, freeze, reject, rate limits, and microsecond-safe trim offsets.
- Normalize equivalent SRT and ASS fixtures and assert identical cue ranges/text.
- Mutate both the source Template and returned views after compilation; assert compiled state is unchanged.
- Test fast and strict fingerprints, explicit base-dir behavior, context cancellation, transition compatibility, shortest-output layer/BGM bounds, and silent video.
- Include one real FFprobe integration test using media dynamically generated in a temporary directory; skip only when the binaries are unavailable.
- Run `go test -race ./ffprobe ./ffvmix`, `go vet ./ffprobe ./ffvmix`, `staticcheck ./ffprobe ./ffvmix`, and verify `go list -deps ./ffvmix` contains no renderer/filter package.

## 7. Wrong vs Correct

### Wrong

```go
// CWD-dependent, repeats I/O during generation, and loses precision.
duration := time.Duration(float32(stream.Duration) * float32(time.Second))
path, _ := filepath.Abs(candidate.Path)
project := generator.NextAfterRunningFFprobe(path)
```

### Correct

```go
compiled, err := ffvmix.Compile(ctx, template,
    ffvmix.WithBaseDir(templateDirectory),
    ffvmix.WithProbeConcurrency(4),
    ffvmix.WithStrictFingerprint(false),
)
if err != nil {
    return err
}

// The generator performs pure reads from compiled slots, joins, assets,
// normalized cues, and adaptation plans.
slots := compiled.Slots()
```

The correct form makes I/O finite, deterministic, testable, and isolated from combination generation.
