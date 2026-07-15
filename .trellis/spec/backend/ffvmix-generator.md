# FFVMix Lazy Generator

## 1. Scope / Trigger

Use this contract when enumerating a `CompiledTemplate`, adding a generation
constraint, or constructing an `ffcut.Project` from FFVMix selections.

Generation is a pure in-memory phase:

```text
CompiledTemplate
    -> lazy unique tuple
    -> CandidateView + accepted HistoryView
    -> constraints
    -> validated ffcut.Project
```

It must not stat files, run FFprobe, parse subtitles, import FFmpeg filters, or
reinterpret template duration policies. Those responsibilities belong to
`ffvmix.Compile`.

## 2. Signatures

```go
func ffvmix.NewGenerator(
    compiled *ffvmix.CompiledTemplate,
    options ...ffvmix.GeneratorOption,
) (*ffvmix.Generator, error)

func ffvmix.WithSeed(seed uint64) ffvmix.GeneratorOption
func ffvmix.WithSearchBudget(maximum uint64) ffvmix.GeneratorOption
func ffvmix.WithConstraint(value ffvmix.Constraint) ffvmix.GeneratorOption
func ffvmix.WithConstraintFunc(
    id string,
    fingerprint string,
    check ffvmix.ConstraintFunc,
) ffvmix.GeneratorOption

func (g *ffvmix.Generator) Seed() uint64
func (g *ffvmix.Generator) Stats() ffvmix.GenerationStats
func (g *ffvmix.Generator) Next(ctx context.Context) (ffvmix.GenerationResult, error)

type Constraint interface {
    ID() string
    Fingerprint() string
    Check(CandidateView, HistoryView) (Decision, error)
}
```

`GenerationResult.Status` is exactly one of `Yielded`, `Exhausted`, or
`BudgetExceeded`. `Project` is non-nil only for `Yielded`.

## 3. Contracts

### Enumeration and seed

- Dimensions are every slot video pool, every join transition pool, and one
  BGM pool dimension only when the compiled template has BGMs.
- Each dimension receives a deterministic weighted permutation without
  replacement. Weight biases earlier positions; it never creates repeated
  samples or removes an option.
- A seeded diagonal min-heap expands index tuples lazily. A visited tuple set
  makes every exact combination reachable at most once without allocating the
  Cartesian product up front.
- Random trim is derived from seed, template fingerprint, slot ID, and video
  ID. It is not another dimension.
- When the caller omits a seed, the generator obtains one from cryptographic
  system entropy and exposes the actual value through `Seed()` and Project
  metadata.

### Search and state

- `Next` scans at most its configured budget. If unvisited tuples remain, a
  rejected budget window returns `BudgetExceeded`; only an empty enumerator
  returns `Exhausted`.
- Engine-infeasible and constraint-rejected tuples advance enumeration and
  update rejection statistics. Constraint errors abort the call after the
  attempt but do not update accepted history or rejection counters.
- Filtering is greedy in seeded order. Rejected tuples are not reconsidered;
  there is no backtracking or maximum-cardinality guarantee.
- Concurrent `Next` calls fail with `ErrConcurrentNext`; they are not silently
  serialized because scheduler order would weaken reproducibility.

### Constraint boundary

- `ffvmix/constraints` owns `CandidateView`, `HistoryView`, the plugin
  interface, and built-ins. The root package aliases the public contract. This
  direction avoids a root/subpackage import cycle.
- Views own their slices and pointers and expose copy-returning getters.
  Constraints can inspect selections but cannot commit generator history.
- Template built-ins run in template order, followed by custom constraints in
  option order. IDs are unique and every constraint supplies a non-empty stable
  configuration fingerprint for Project provenance.
- `MaxSimilarity` compares a candidate with each accepted output:

  ```text
  similarity =
      sum(source-range intersection for the same slot and normalized path)
      / sum(max(left source duration, right source duration) per slot)
  ```

  Equality with `maximum` is accepted; only a greater value is rejected. BGM,
  transitions, and fixed layers are excluded.
- `MaxVideoAssetUses` counts selection occurrences by normalized video path,
  including multiple ranges or slots from the same file. `MaxBGMUses` has a
  separate counter by normalized BGM path.

### Project construction

- Constraints run before the full Project is allocated. Accepted selections
  become clips whose source windows use the precompiled adaptation plan.
- Clip zero starts at zero. Each later clip starts at the previous clip end
  minus the selected transition duration. A transition range is exactly that
  overlap.
- A selected BGM starts at its compiled timeline start and plays to output end;
  a non-looping track is capped by its source range.
- `full_output`, `slot`, and `absolute` layer timing becomes an absolute
  `ffcut.TimeRange`. Subtitle cue ranges are shifted by the resolved layer
  start. Layer slice order remains template render order.
- Project construction and `Project.Validate()` must succeed before accepted
  history is committed.
- Metadata records template fingerprint, actual seed, combination fingerprint,
  all selections with asset fingerprints, and every constraint fingerprint.
  Synthetic Project and BGM-dimension IDs deterministically avoid persisted ID
  collisions.

## 4. Validation & Error Matrix

| Condition | Required result |
|-----------|-----------------|
| Nil/incomplete `CompiledTemplate`, zero budget, empty dimension, invalid weight | Error matching `ErrInvalidGenerator` |
| Nil custom constraint, empty/duplicate constraint ID, empty fingerprint | Error matching `ErrInvalidGenerator` |
| Context canceled before the next attempt | Return the context error and do not advance |
| A second overlapping `Next` call | `ErrConcurrentNext`; active call continues unchanged |
| Infeasible video plan | Reject with `ReasonInfeasibleVideo` and advance |
| Transition incompatible with selected adjacent videos | Reject with `ReasonIncompatibleTransition` and advance |
| Constraint rejection with stable reason | Increment rejection stats; accepted history is unchanged |
| Constraint returns an error | Error matching `ErrConstraintCheck` and the plugin cause; history is unchanged |
| Constraint rejects without a reason | `ErrConstraintCheck` |
| Timeline arithmetic or Project validation fails | Error matching `ErrProjectBuild`, preserving the underlying cause |
| Budget ends while tuples remain | `BudgetExceeded`, never `Exhausted` |
| Enumerator contains no tuple | `Exhausted`, including repeated later calls |

Callers use `errors.Is`; they must not parse formatted error strings.

## 5. Good / Base / Bad Cases

- Good: a two-slot, two-transition, two-BGM space yields every valid exact
  combination once, although its order changes with the seed.
- Good: a high-weight option usually appears early but remains a single option
  in the terminal set.
- Good: a one-attempt budget rejects one tuple, returns `BudgetExceeded`, and
  resumes from the next tuple on the following call.
- Base: one natural-duration slot, no join, BGM, layer, or constraint yields one
  Project followed by `Exhausted`.
- Bad: treating a rejected budget window as proof that the Cartesian space is
  exhausted.
- Bad: updating a plugin-owned counter during `Check`; a later constraint error
  would make accepted history and plugin state disagree.
- Bad: constructing a Project before cheap engine and history constraints run.

## 6. Tests Required

- Exhaust a small Cartesian product and assert exact count, unique combination
  fingerprints, stable same-seed order, and seed-independent terminal set.
- Across many seeds, assert a heavier option is earlier statistically while all
  options remain present.
- Cover budget resume, pre-attempt cancellation, repeated exhaustion, constraint
  rejection, constraint error, and overlapping `Next` calls under `-race`.
- Boundary-test `MaxSimilarity`, `MaxVideoAssetUses`, and `MaxBGMUses`, including
  equality acceptance and independent video/BGM counters.
- Assert engine rejection reason counts for infeasible videos and incompatible
  transitions.
- Assert random trim is same-seed stable, different-seed variable, microsecond
  aligned, and within its compiled maximum.
- Assert clip starts, fade overlap, silent original audio, BGM range, layer
  anchors, subtitle cue shifts, metadata, and Project validation.
- Compile a real template through a fake prober, generate a Project, round-trip
  it through `ffcut.Marshal`/`Unmarshal`, and assert generation performs no new
  probes.
- Run `go test -race -cover ./ffvmix ./ffvmix/constraints`, `go vet` and
  `staticcheck` for both packages, and verify `go list -deps ./ffvmix` contains
  no `ffmpeg` or `ffcut/fusion` package.

## 7. Wrong vs Correct

### Wrong

```go
// Allocates the full product, samples with replacement, and conflates a scan
// limit with terminal exhaustion.
projects := buildEveryProject(compiled)
for attempts := 0; attempts < 100; attempts++ {
    project := weightedRandomChoice(projects)
    if constraint(project) {
        return project, nil
    }
}
return nil, io.EOF
```

### Correct

```go
generator, err := ffvmix.NewGenerator(compiled,
    ffvmix.WithSeed(42),
    ffvmix.WithSearchBudget(100),
    ffvmix.WithConstraintFunc("campaign", "campaign-v1", checkCampaign),
)
if err != nil {
    return err
}

result, err := generator.Next(ctx)
if err != nil {
    return err
}
switch result.Status {
case ffvmix.Yielded:
    consume(result.Project)
case ffvmix.BudgetExceeded:
    // Call Next again to continue from preserved progress.
case ffvmix.Exhausted:
    // The generator proved no tuple remains.
}
```

The correct form preserves uniqueness, reproducibility, bounded work, and an
atomic accepted-history boundary.
