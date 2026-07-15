# FFVMix Generator Design

## Interface

```go
generator, err := ffvmix.NewGenerator(compiled, options...)
seed := generator.Seed()

result, err := generator.Next(ctx)
switch result.Status {
case Yielded:
    project := result.Project
case Exhausted:
case BudgetExceeded:
}
```

`Next` is stateful and serial. It commits state only after all constraints accept and Project construction succeeds.

## Enumerator

Each combination dimension first receives a deterministic weighted option order derived from the generator seed. A heap-based diagonal Cartesian traversal walks the resulting index tuples lazily, rather than nesting loops that bias the first dimension or allocating the full product.

A visited tuple set guarantees uniqueness. The engine combination fingerprint includes selected source IDs, transition IDs and BGM ID. Random trim offsets are derived values and do not create dimensions.

Weights are an ordering bias. They do not mean that an option can be sampled repeatedly, and the no-constraint terminal set is independent of weights and seed.

## Processing Pipeline

For each tuple:

1. Build a cheap CandidateView from compiled metadata.
2. Reject engine-infeasible source/transition combinations.
3. Run built-in constraints in template order.
4. Run injected custom constraints in option order.
5. Resolve slot start times and transition overlaps.
6. Resolve semantic layer anchors to absolute Project ranges.
7. Construct and validate `ffcut.Project`.
8. Commit accepted history and return Yielded.

Rejected candidates advance the enumerator but never mutate accepted history.

## Constraints

```go
type Constraint interface {
    ID() string
    Fingerprint() string
    Check(CandidateView, HistoryView) (Decision, error)
}
```

CandidateView exposes immutable selections, normalized paths, source/timeline durations, transitions, BGM and fingerprints. It does not expose a full Project. HistoryView exposes accepted summaries only.

The view contract and built-in implementations live in the independent `ffvmix/constraints` package; root `ffvmix` aliases those types. This keeps the dependency direction `ffvmix -> ffvmix/constraints -> ffcut` and avoids a root/subpackage import cycle.

`MaxSimilarity` is symmetric per slot: the numerator is the source-range intersection only when slot and normalized path match; the denominator sums the larger source duration for each aligned slot. Equality with the configured maximum is accepted. Video-path occurrences and BGM-path occurrences use independent counters.

`Decision` contains Accept/Reject and a stable reason code. Generator stats aggregate attempts and rejections by reason.

## Search Budget

Every `Next` invocation scans at most a configured number of tuples. If no Project is accepted before that limit, it returns BudgetExceeded with current statistics and preserves the advanced enumerator. Exhausted is returned only after the enumerator proves that no tuple remains.

Context cancellation returns the context error and leaves all fully processed tuple advances intact; no half-committed accepted history is possible.

## Project Compilation

The compiler turns semantic selections into absolute timeline data:

- fixed and natural slot durations;
- source windows, playback rates, loops and freeze tails;
- overlap-aware clip starts and transitions;
- original-audio ranges and gains;
- selected BGM loop/fade plan;
- ordered global layers with resolved ranges and canvas geometry;
- provenance metadata.

The generator depends only on `ffcut` protocol types and compiled metadata, never FFmpeg filters.

## Concurrency

The public contract states that concurrent Next calls are unsupported. The implementation uses a lightweight re-entry guard to fail clearly rather than allowing silent data races; it does not serialize callers because scheduling order would undermine reproducibility.

Synthetic BGM-dimension and Project IDs are derived deterministically and extended when necessary so raw persisted IDs cannot make otherwise valid generated Projects fail uniqueness validation.
