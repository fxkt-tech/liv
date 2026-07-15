# FFVMix Generator Implementation Plan

Prerequisites: `07-15-ffcut-project-v2` and `07-15-ffvmix-template-compiler` are complete.

1. Define Generator options, seed handling, result statuses, statistics and re-entry guard.
2. Implement deterministic weighted ordering for one dimension with tests.
3. Implement lazy diagonal Cartesian traversal and uniqueness tests.
4. Add CandidateView, HistoryView, Decision and Constraint contracts.
5. Implement template-configured built-in constraints.
6. Implement custom constraint composition and error atomicity.
7. Add feasibility filtering and stable rejection reasons.
8. Implement absolute timeline scheduling with transition overlap.
9. Compile original audio, selected BGM and semantic global layers into Project.
10. Add provenance metadata and Project validation before commit.
11. Add search-budget, cancellation, exhaustion and resume tests.
12. Add property tests for determinism, uniqueness and complete small-space coverage.
13. Run:
    - `gofmt` on changed Go files
    - `go test ./ffvmix ./ffvmix/constraints`
    - `go vet ./ffvmix ./ffvmix/constraints`

## Rollback Points

- Enumerator and constraint contracts land before Project construction.
- Built-in constraints remain separate from engine invariants.
- No FFmpeg imports or renderer code are allowed in this child.
