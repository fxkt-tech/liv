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
