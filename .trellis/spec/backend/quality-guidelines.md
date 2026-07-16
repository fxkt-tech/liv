# Quality Guidelines

> Verification and review standards derived from the current Go codebase.

## Baseline Gate

The module targets Go 1.25 (`go.mod`). Before committing a change:

```bash
gofmt -w <changed-go-files>
go vet ./...
go test ./...
```

The repository Makefile provides the verbose coverage form:

```bash
make test  # go test -v ./... -cover
```

No third-party linter or CI workflow is configured in this repository. Do not
claim a tool-specific rule that the project does not enforce.

Documentation-only changes still require template-marker/link checks and should
run the Go test suite when they change development rules.

### Known non-hermetic test baseline

`ffcut/fusion/track_test.go` currently mixes one serialization test with four
manual FFmpeg examples: `TestExec`, `TestExec2`, `TestSpeed`, and `TestVmix`.
Those four tests depend on `in.mp4` or absolute files under a developer desktop
and can write `out_test*.mp4`. On a clean machine, `go test ./...` therefore
fails with missing-input errors even when all hermetic packages pass.

Until those manual examples are converted to generated `t.TempDir()` fixtures:

```bash
# Still run this first and confirm any failure is understood.
go test ./...

# Verify every package outside the known non-hermetic package.
go list ./... | rg -v '/ffcut/fusion$' | xargs go test

# Verify the hermetic serialization test in that package.
go test ./ffcut/fusion -run '^TestExport$'
```

Do not broaden this exception or copy the manual-test pattern. Any change to
`ffcut/fusion` must address the affected tests explicitly; unrelated changes
may record the exact baseline failure after the two hermetic commands pass.

## Testing Patterns

Tests are co-located with the owning package as `*_test.go` and use Go's
standard `testing` package.

- Use table-driven subtests for input matrices. The primary example is
  `ffcut/validate_test.go:TestProjectValidateRejectsInvalidValues`.
- Use `t.TempDir()` for filesystem fixtures and `t.Helper()` for fixture
  builders (`ffvmix/compile_test.go`, `ffprobe/ffprobe_test.go`).
- Put external tools behind a narrow private interface when core behavior can
  be tested without process I/O. `ffvmix.mediaProber` and `fakeMediaProber` are
  the reference.
- If a true FFmpeg/FFprobe integration test is needed, generate its media in a
  temporary directory and skip only when the binary is absent, as
  `ffprobe/ffprobe_test.go` does.
- Assert error identity with `errors.Is` / `errors.As`, including wrapped
  causes. Do not lock tests to a full human-readable message unless the string
  is itself a protocol.

## Contract Changes

For JSON protocol or template changes, test the whole boundary:

1. construct a valid value;
2. validate it;
3. marshal and unmarshal it;
4. reject unknown fields and trailing JSON;
5. compare the round-tripped value;
6. add targeted invalid cases with stable field paths.

References: `ffcut/marshal_test.go`, `ffcut/validate_test.go`, and
`ffvmix/template_test.go`.

For FFVMix generation, also test deterministic ordering for a fixed seed,
uniqueness of outputs, constraint history, cancellation, and absence of
generation-time I/O (`ffvmix/generator_test.go`). Follow the dedicated FFcut
and FFVMix specs in this directory for their full contracts.

## Required Implementation Patterns

- Pass `context.Context` into blocking external work and honor cancellation.
- Keep persisted input validation strict and centralized at its owner.
- Return copies or immutable views when callers must not mutate engine state;
  `ffvmix.Generator.Stats` and compiled-template accessors are examples.
- Keep generated protocol output valid by calling the owning validator before
  returning or encoding it.
- Use functional options where the package already uses them instead of adding
  parallel constructors with growing parameter lists.
- Search for existing constants, codecs, helpers, and validators before adding
  another definition.

## Patterns Not to Copy

- Unconditional debug printing from library code.
- Panics for invalid caller input.
- String-only errors when callers need a stable classification.
- Real filesystem or FFprobe work during FFVMix lazy generation.
- Permissive decoding of persisted FFcut or FFVMix JSON.
- Cross-package shortcuts that move FFcut protocol invariants into FFmpeg or
  FFVMix implementation code.
- Tests that require a developer's fixed local media path or leave artifacts
  outside `t.TempDir()`.

Some older packages predate the stricter FFcut/FFVMix contracts. Their behavior
is evidence of compatibility constraints, not an automatic template for new
code. Prefer the package-local recent pattern and avoid unrelated cleanup in a
feature change.

## Review Checklist

- Does the change live in the package that owns its invariant?
- Are success, invalid input, boundary values, and cancellation covered where
  applicable?
- Are errors classifiable and wrapped without losing their cause?
- Does any serialization or public API change update validation, round-trip
  tests, examples, and the relevant Trellis spec?
- Are generated results deterministic and isolated from caller mutation where
  promised?
- Do `gofmt` and `go vet ./...` pass?
- Does `go test ./...` pass, or is its only failure the exact documented
  `ffcut/fusion` baseline with both hermetic verification commands passing?
