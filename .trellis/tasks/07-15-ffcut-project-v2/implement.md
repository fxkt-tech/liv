# FFcut Project v2 Implementation Plan

1. Replace the empty root `ffcut` facade with version, ID, duration, range and enum primitives.
2. Add canvas, background, local source and media fingerprint types.
3. Add sequence clips, playback behavior, transitions and audio tracks.
4. Add ordered subtitle/image layers and px/percent geometry.
5. Add Project metadata and provenance selection records.
6. Implement aggregate validation with field-path and sentinel errors.
7. Implement validated JSON marshal/unmarshal without mutation.
8. Add table tests for every validation category.
9. Add JSON golden/round-trip and non-mutation tests.
10. Run:
    - `gofmt` on changed Go files
    - `go test ./ffcut`
    - `go vet ./ffcut`

## Risk and Rollback

- Do not edit `ffcut/fusion` in this child.
- Do not introduce FFmpeg or FFprobe imports.
- Keep each major type group in its own file so the entire v2 protocol can be reverted without touching legacy packages.
