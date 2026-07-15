# FFVMix Parent Execution Plan

This parent task coordinates three independently verifiable child tasks. It has no direct production-code implementation.

## Ordered Children

1. `07-15-ffcut-project-v2`
   - Establish and review the shared timeline contract.
   - Must finish before either `ffvmix` child starts.
2. `07-15-ffvmix-template-compiler`
   - Build persisted templates, local-media compilation and normalized immutable state.
   - Depends on the Project v2 types.
3. `07-15-ffvmix-generator`
   - Build lazy weighted traversal, constraints and Project compilation.
   - Depends on both earlier children.

## Parent Integration Review

- Verify the dependency graph contains no `ffvmix -> ffmpeg/filter` edge.
- Run targeted tests for `ffcut`, `ffprobe` and `ffvmix`.
- Run full repository tests and separate new failures from known legacy fixture failures.
- Verify all child acceptance criteria map back to the parent PRD.
- Verify example code uses repository-controlled or generated media fixtures.
- Review whether final decisions belong in `.trellis/spec/backend/` before archiving.

## Rollback Points

- Child 1 can be reverted without touching legacy `ffcut/fusion`.
- Child 2 must not expose partially probed mutable templates; revert if the immutable compile contract cannot be maintained.
- Child 3 can be removed without changing persisted template or Project formats.

## Parent Completion Gate

- All three children are checked and archived.
- Cross-child tests pass.
- Parent PRD has no unmet acceptance criteria.
- Renderer remains explicitly out of scope.

## Integration Review — 2026-07-15

- [x] All three child tasks are archived with every child acceptance criterion checked.
- [x] `go test -count=1 -race -cover ./ffcut ./ffprobe ./ffvmix ./ffvmix/constraints` passes.
- [x] `go vet` and `staticcheck` pass for `ffcut`, `ffprobe`, `ffvmix`, and `ffvmix/constraints`.
- [x] `ffcut` has no dependency on `ffprobe`, `ffmpeg`, or `ffcut/fusion`; `ffvmix` has no dependency on `ffmpeg` or `ffcut/fusion`.
- [x] Compile-to-generator integration proves no generation-time probing and Project v2 semantic JSON round-trip.
- [x] Generator integration covers natural, speed-up, slow-down, loop, freeze, trim, all background variants, transitions, original audio, BGM, and global layers.
- [x] New tests contain no developer-home or desktop media paths; real media is generated in temporary directories.
- [x] Full `go test -count=1 ./...` has no new failures. The only failures are the four pre-existing non-hermetic `ffcut/fusion` tests (`TestExec`, `TestExec2`, `TestSpeed`, `TestVmix`) that require `in.mp4` or files under `/Users/justyer/Desktop`.
- [x] Project v2, template compiler, and lazy generator contracts are captured under `.trellis/spec/backend/`.
- [x] FFmpeg rendering and migration of legacy `ffcut/fusion` remain explicitly out of scope.
