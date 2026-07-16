# Logging Guidelines

> Observable output in a library that does not own application logging.

## Current Logging Model

Liv has no logging dependency, structured log schema, or application log-level
contract. Core packages return errors and let the embedding application decide
where and how to log them.

Do not introduce a logger or emit routine progress messages from library code
as part of an unrelated feature.

## Existing Diagnostic Output

`ffmpeg` and `ffprobe` have explicit debug/dry-run command output:

- `ffmpeg.WithDebug(true)` prints the command before execution.
- `ffmpeg.WithDry(true)` prints the command and skips execution.
- `ffprobe.WithDebug(true)` prints the command before execution.
- `ffprobe.FFprobe.Sentence` records the executed command string.

These paths use `fmt.Println(strings.Join(...))` in `ffmpeg/ffmpeg.go` and
`ffprobe/ffprobe.go`. They print to standard output and are command previews,
not structured logs.

`ffmpeg.WithLogLevel` controls FFmpeg's own `-v` argument. It is not a Go
application logger and must not be described as one.

## Rules for New Code

- Return errors from library operations; the caller owns operational logging.
- Emit command text only behind an explicit debug or dry-run option.
- Keep normal protocol validation, template compilation, and generation silent.
- Never add unconditional `fmt.Print*` calls for progress or debugging.
- Be aware that command previews can contain local paths, URLs, user-agent
  values, metadata, or other caller-provided arguments. Callers must not enable
  debug output where those values are sensitive.
- A future need for structured observability requires an explicit public API
  design (for example, a callback or injected interface), documented event
  ownership, and tests. Do not add a package-global logger.

`transcode.go` currently contains an unconditional
`fmt.Println("处理filter")`. This is legacy debug output and an explicit
exception; do not copy it as a convention.

## Log Levels and Fields

No application log levels or mandatory structured fields are established.
Define neither until the project adds an actual logging abstraction. Downstream
services may attach their own operation IDs, media identifiers, durations, and
error classifications when they log calls into Liv.

## Review Checklist

- Is normal library execution silent?
- Is diagnostic output opt-in and deterministic?
- Could printed command arguments expose caller data?
- Is a returned error carrying information that was otherwise only printed?
