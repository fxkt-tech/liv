# FFVMix Technical Design

## 1. Design Position

`ffvmix` is a deterministic, lazy compiler from a persisted mix template to typed `ffcut.Project` values. It is not a renderer and must not know FFmpeg filter syntax.

The clean seam is the typed timeline protocol:

```text
Template JSON
    -> ffvmix.Compile(ctx)
    -> immutable CompiledTemplate
    -> Generator.Next(ctx)
    -> ffcut.Project v2
    -> future ffcut/renderer
```

Dependencies only point inward:

```text
ffvmix -> ffcut
ffvmix -> ffprobe
future ffcut/renderer -> ffcut
```

The existing `ffcut/fusion` package is frozen during this work. New behavior must not be added to its monolithic `Export` path.

## 2. Package Boundaries

```text
ffcut/
  project.go       Project and canvas
  timeline.go      clips, joins, audio
  layer.go         subtitle/image layers
  time.go          microsecond protocol time
  validate.go      semantic validation
  marshal.go       validated JSON codec

ffvmix/
  template.go      persisted aggregate, slots, sources, joins and BGM
  adaptation.go    duration and fit policy planning
  layer.go         semantic layer anchors
  compile.go       validation/probing/normalization
  compiled.go      immutable compiled views
  generator.go     stateful iterator
  enumerator.go    deterministic weighted traversal
  constraint.go    root aliases for the plugin contract
  constraints/     immutable views and built-in constraints
  project.go       absolute ffcut.Project construction
  errors.go        typed, contextual errors
```

## 3. Domain Model

### Template

The template is data-only and JSON serializable. Runtime functions are never stored in it.

```go
type Template struct {
    ID            ID
    SchemaVersion int
    Canvas        CanvasSpec
    Background    BackgroundSpec
    Defaults      SlotDefaults
    Slots         []Slot
    Joins         []Join
    BGMs          []Weighted[AudioSource]
    Layers        []LayerSpec
    Constraints   []ConstraintSpec
}
```

Constructors generate stable IDs once. IDs are serialized; decoding raw JSON with missing IDs is an error. Slot-local time anchors and join references use those IDs.

### Slot

A slot is one required position in the primary sequence. Each generated candidate selects exactly one source from each slot.

An unset target duration uses the selected source range duration at rate 1. A fixed target duration chooses one explicit overflow or underflow policy. Policies never silently fall back to another policy.

### Join

A join is the edge between adjacent slots. It owns one or more weighted transition candidates, including `cut`. Transition overlap is part of both adjacent slot durations, so total duration is:

```text
sum(slot durations) - sum(selected transition durations)
```

### Layers

Template layers use semantic time anchors. Project compilation resolves them to absolute ranges after all slot durations and transitions are known. Spatial values are tagged `px` or `percent`; untagged floats are invalid.

## 4. ffcut.Project v2

The Project is execution-ready and contains no slot-anchor semantics. It contains:

- versioned canvas, FPS and background;
- a primary video sequence with source and timeline ranges;
- explicit playback behavior: rate, loop, freeze-last-frame;
- explicit transitions between clip IDs;
- separate original-audio and BGM tracks;
- ordered, absolute-time subtitle/image layers;
- provenance metadata.

Protocol time is a named `int64` microsecond type. JSON unions use a discriminator plus a typed payload; `map[string]any` and unrestricted `interface{}` payloads are forbidden.

Marshal and unmarshal validate the complete Project and return errors. Serialization must not mutate the Project.

## 5. Compile Phase

`Compile` performs all I/O and returns an immutable value:

1. Validate template shape, generated IDs, slot order and references.
2. Resolve relative paths against an explicit base directory.
3. De-duplicate local paths and probe them with bounded concurrency.
4. Use the first video/audio stream; missing video is invalid, missing audio means silence.
5. Validate source ranges and collect display dimensions/duration.
6. Record size + mtime fingerprints, optionally SHA-256.
7. Parse local SRT/ASS sources into normalized cues.
8. Precompute slot-source adaptation plans and join compatibility.
9. Aggregate all validation errors instead of failing at the first bad asset.

No probing occurs in `Next`. The existing ffprobe duration fields must be upgraded from `float32` before they feed microsecond protocol time.

## 6. Duration Adaptation

For a fixed slot duration `D` and selected source range duration `S`:

- `S == D`: rate 1, no fill.
- `S > D`:
  - `speed_up`: rate `S/D`, valid only within configured maximum;
  - `trim`: select a `D` source window at rate 1;
  - `reject`: combination is infeasible.
- `S < D`:
  - `slow_down`: rate `S/D`, valid only within configured minimum;
  - `loop`: repeat and trim the final repetition;
  - `freeze`: play once, hold the final frame, and emit silence for the remainder of original audio;
  - `reject`: combination is infeasible.

Random trim offset is a pure function of seed, template fingerprint, slot ID and source ID. It does not create an additional combination dimension.

## 7. Lazy Enumeration

Combination dimensions are:

1. one source per slot;
2. one transition per join;
3. zero or one BGM choice from the configured pool.

Deterministic weighted ordering is implemented without materializing the Cartesian product. Each dimension receives a seeded weighted ordering, then a heap-based diagonal traversal yields tuples lazily and exactly once. Weight is documented as an order bias, not a promise of repeated statistical sampling.

The generator applies checks in this order:

1. engine invariants and exact-combination uniqueness;
2. source/transition feasibility;
3. template built-ins followed by custom constraints against accepted history;
4. Project construction;
5. history commit.

Rejected tuples are never reconsidered. The result is greedy and order-dependent; no maximum-cardinality claim is made.

`Next` has explicit `Yielded`, `Exhausted` and `BudgetExceeded` states. Context controls cancellation; a scan budget prevents a single call from traversing an unbounded rejected region. Concurrent `Next` consumption is unsupported, and an atomic re-entry guard returns `ErrConcurrentNext` rather than silently serializing callers or allowing a data race.

## 8. Constraint Contract

```go
type Constraint interface {
    ID() string
    Fingerprint() string
    Check(candidate CandidateView, history HistoryView) (Decision, error)
}
```

Views are read-only. Constraints cannot commit counters or mutate history. An error aborts the call; a rejection includes a stable reason code for statistics.

Built-ins:

- `MaxSimilarity`: duration-weighted overlap where the same slot selects the same source; fixed layers and BGM are excluded.
- `MaxVideoAssetUses`: counts normalized local paths, so ranges from one file share the quota.
- `MaxBGMUses`: independent quota.

Built-in specs are persisted in the template. Custom Go implementations are injected when the generator is created and must contribute a stable configuration fingerprint for provenance.

## 9. Provenance and Reproducibility

Every Project records:

- template content fingerprint;
- actual generator seed;
- combination fingerprint;
- slot/source, join/transition and BGM selections;
- local file fingerprints;
- built-in and custom constraint fingerprints.

Strict reproducibility can opt into SHA-256 input fingerprints. Default probing uses size and mtime for lower startup cost.

## 10. Compatibility and Rollout

- There is no compatibility requirement for the existing fusion JSON.
- Land the pure Project v2 module first.
- Land template compilation only after the Project contract is reviewed.
- Land the generator last, behind tests that prove determinism and uniqueness.
- Do not delete `ffcut/fusion` until users have migrated; freezing it keeps rollback trivial.
- FFmpeg rendering is a separate future task and is not required for this parent task.
