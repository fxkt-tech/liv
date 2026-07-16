# Error Handling

> Error contracts for library boundaries, validation, codecs, and media tools.

## Core Contract

Liv is a library. Functions return errors to their caller; they do not log and
swallow failures. Preserve both machine-checkable causes and enough local
context to identify the failing field, file, operation, or external command.

## Established Error Shapes

### Sentinel plus wrapping

Define stable package-level sentinels when callers need to classify a failure,
then wrap them with `%w`:

```go
var ErrInvalidGenerator = errors.New("ffvmix: invalid generator")

return nil, fmt.Errorf("%w: compiled template is required", ErrInvalidGenerator)
```

References: `ffcut/errors.go`, `ffprobe/duration.go`,
`ffvmix/generator.go`. Tests use `errors.Is`, not string equality; see
`ffcut/time_test.go` and `ffvmix/generator_test.go`.

### Aggregated validation issues

When one pass can report several independent input defects, collect them in a
typed error rather than returning only the first:

- `ffcut.ValidationError` contains field paths and implements `Is` and
  multi-error `Unwrap` (`ffcut/errors.go`, `ffcut/validate.go`).
- `ffvmix.CompileError` contains an `IssueCode`, template path, resolved local
  path, message, and optional cause (`ffvmix/errors.go`, `ffvmix/compile.go`).

Use stable field paths such as `slots[0].videos[1].path`. Callers and tests rely
on them to locate invalid configuration.

### Codec errors

Persisted JSON boundaries reject unknown fields and trailing values before
validation. `ffcut.CodecError` and `ffvmix.TemplateCodecError` retain the
operation (`encode` or `decode`) and unwrap their cause. Follow
`ffcut/marshal.go` and `ffvmix/codec.go` for new persisted formats.

### External command failures

Use `exec.CommandContext`, propagate cancellation, and include trimmed stderr
along with the original execution error. `ffprobe.Run` is the current reference:

```go
return fmt.Errorf("ffprobe failed: %w: %s", err,
    strings.TrimSpace(string(output)))
```

`ffmpeg.Run` currently replaces the process error with `errors.New(stderr)`;
that is legacy behavior, not a pattern to copy into new adapters.

## Boundary Rules

- Validate at the boundary that owns the contract: template validation before
  compilation, compiled-state validation before generation, and FFcut project
  validation before encoding.
- Wrap an error when adding useful operation or identity context. Use `%w` for
  causes that callers may inspect.
- Preserve `context.Canceled` and `context.DeadlineExceeded` through wrapping.
- Do not compare error messages when `errors.Is` or `errors.As` can express the
  contract.
- Do not expose partial compiled or protocol output on validation failure.
- Nil public receivers or required arguments return a classified error rather
  than panic; see `ffcut.Project.Validate` and `ffvmix.Generator.Next`.

## HTTP/API Responses

There is no HTTP server or standard API error envelope in this repository.
Transport adapters added by downstream applications must translate Liv errors
outside this module; do not add HTTP status codes to `ffcut`, `ffprobe`,
`ffmpeg`, or `ffvmix` errors.

## Review Checklist

- Can callers classify the error with `errors.Is` / `errors.As`?
- Does the message name the operation and relevant field/path without losing
  the cause?
- Are all independently detectable validation issues returned together?
- Does cancellation survive the full call chain?
- Is there a negative or regression test for the error contract?
