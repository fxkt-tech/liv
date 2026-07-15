# FFVMix Template Compiler Implementation Plan

Prerequisite: `07-15-ffcut-project-v2` is complete.

1. Add serializable Template, Slot, Join, BGM, Layer and built-in constraint spec types.
2. Add constructors that generate and persist IDs, plus JSON codecs that do not synthesize missing IDs.
3. Add structural validation and aggregate issue types.
4. Upgrade FFprobe duration precision with regression tests.
5. Add internal prober seam and bounded, de-duplicated local probing.
6. Add base-directory path resolution and fast/strict file fingerprints.
7. Add source-range and first-stream validation; model missing audio as silence.
8. Add SRT and ASS parsers that normalize to compiled cues.
9. Add duration adaptation and transition compatibility planning.
10. Build immutable `CompiledTemplate` and mutation-isolation tests.
11. Add fake-prober unit tests and one hermetic real-FFprobe integration test.
12. Run:
    - `gofmt` on changed Go files
    - `go test ./ffprobe ./ffvmix`
    - `go vet ./ffprobe ./ffvmix`

## Rollback Points

- FFprobe precision change lands with its own tests and can be reverted separately.
- Template JSON types land before filesystem compilation.
- No generator or renderer behavior belongs in this child.
