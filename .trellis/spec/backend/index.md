# Backend Development Guidelines

> Coding contracts for Liv's Go media-processing packages.

## Overview

Liv is a single-module Go library. The specs in this directory describe its
package boundaries, error and diagnostic behavior, verification gate, and the
persisted FFcut/FFVMix contracts. They are derived from the current repository;
there is no HTTP or database application layer.

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | Package ownership, dependency direction, and file placement | Established |
| [Database Status](./database-guidelines.md) | Why persistence conventions are currently not applicable | Not applicable |
| [Error Handling](./error-handling.md) | Sentinels, wrapping, aggregated validation, and command failures | Established |
| [Logging](./logging-guidelines.md) | Silent library behavior and opt-in command diagnostics | Established |
| [Quality](./quality-guidelines.md) | Go verification gate, tests, and review rules | Established |
| [FFcut Project v2](./ffcut-project-v2.md) | Typed timeline protocol and validation boundary | Established |
| [FFcut Renderer](./ffcut-renderer.md) | Supported executable FFmpeg subset and artifact contract | Established |
| [FFVMix Template Compiler](./ffvmix-template-compiler.md) | Persistent templates, local asset compilation, and immutable results | Established |
| [FFVMix Lazy Generator](./ffvmix-generator.md) | Deterministic lazy enumeration, pure constraints, and project construction | Established |

## Pre-Development Checklist

Always read:

- [Directory Structure](./directory-structure.md)
- [Quality](./quality-guidelines.md)

Then read the guides matching the change:

- Errors, validation, codecs, or external commands:
  [Error Handling](./error-handling.md)
- Debug/dry-run output or observability: [Logging](./logging-guidelines.md)
- FFcut wire types, timeline, validation, or JSON:
  [FFcut Project v2](./ffcut-project-v2.md)
- FFcut rendering or FFmpeg artifact output:
  [FFcut Renderer](./ffcut-renderer.md)
- FFVMix template or compilation:
  [FFVMix Template Compiler](./ffvmix-template-compiler.md)
- FFVMix enumeration, constraints, or project construction:
  [FFVMix Lazy Generator](./ffvmix-generator.md)
- Persistence: read [Database Status](./database-guidelines.md), then design and
  document the new boundary before implementation because none exists today.

## Quality Check

For Go changes:

```bash
gofmt -w <changed-go-files>
go vet ./...
go test ./...
```

Use `make test` when verbose package output and coverage are useful. For spec
changes, also verify:

```bash
rg -n "[T]o be filled|[T]o fill|TODO:[[:space:]]*fill" .trellis/spec
```

Any match must be reviewed; finished specs must not contain template prose.
Check that this index lists every backend spec and that code paths cited by the
changed documents still exist.
