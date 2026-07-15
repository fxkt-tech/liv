# Current Code Assessment

## Scope

Read-only inspection of the current `ffcut`, `ffprobe` and supporting FFmpeg builder code before designing FFVMix.

## Evidence

- `ffcut/ffcut.go:5` exposes an empty `FFcut` type, so the root package is available for a new typed Project contract.
- `ffcut/fusion/template.go:9` ignores caller options and `FromProto` at line 22 returns success without parsing.
- `ffcut/fusion/track.go:140` combines validation, input construction, timeline compilation, filters, output configuration and process execution in one method.
- `ffcut/fusion/track.go:175` dereferences an optional source section before checking item kind, which makes the current shape unsafe for generalized sources.
- `ffcut/fusion/track.go:306` passes an absolute end time to an input helper whose second parameter is duration.
- `ffcut/fusion/track.go:332` silently ignores title and subtitle tracks during local export.
- `ffcut/fusion/track.go:442` stores track errors internally, while `TrackData.AddTrack` at line 101 does not propagate them.
- `ffcut/fusion/track_test.go:76` and later tests depend on local files such as `in.mp4` and developer Desktop paths. `go test ./ffcut/...` currently fails when those files are absent.
- `ffprobe/model.go:15` and line 42 store duration as `float32`, which is inconsistent with the proposed microsecond Project time.
- `ffprobe/ffprobe.go:128` and line 140 already provide first video/audio stream selection, matching the product decision.

## Consequences

1. Extending `fusion.Export` would preserve the wrong seam and make template generation depend on filter execution details.
2. The shortest stable path is a pure root `ffcut.Project v2`, followed by a separate `ffvmix` compiler and generator.
3. FFprobe duration precision must be improved before media metadata is converted to protocol time.
4. New tests must be hermetic; legacy integration examples cannot serve as regression tests.

## Baseline Commands

- `go vet ./ffcut/...` passes.
- `go test ./ffcut/...` fails in four legacy execution tests because local media files are missing.
