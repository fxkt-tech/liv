# Database Guidelines

> Current status of persistence in this repository.

## Status: Not Applicable

Liv currently has no database layer. `go.mod` contains no SQL, ORM, migration,
key-value store, or database driver dependency, and the repository has no
schema or migration directory. Therefore there are no project conventions for
queries, transactions, table names, indexes, or migrations.

Do not invent those conventions during unrelated work.

## What Exists Instead

The project has serialization and local-media boundaries, but neither is a
database abstraction:

- `ffcut.Marshal` / `ffcut.Unmarshal` encode and validate the FFcut v2 JSON
  protocol (`ffcut/marshal.go`).
- `ffvmix.MarshalTemplate` / `ffvmix.UnmarshalTemplate` encode editable
  templates (`ffvmix/codec.go`).
- `ffvmix.Compile` resolves local files and returns an immutable in-memory
  `CompiledTemplate`; it does not persist templates (`ffvmix/compile.go`).
- `ffprobe.FFprobe` reads metadata from a media file by running `ffprobe`; it is
  an external-process adapter, not a repository (`ffprobe/ffprobe.go`).

Keep persistence concerns out of these protocol and compiler packages unless a
future requirement explicitly changes their contracts.

## If Persistence Is Added

Treat it as a new architectural boundary, not as a utility hidden in `pkg/`.
Before implementation, define:

1. what aggregate is persisted and why JSON files are insufficient;
2. the owning package and its interface to `ffcut` or `ffvmix`;
3. driver/ORM choice, transaction ownership, schema naming, and migrations;
4. tests that do not require a developer's unmanaged database;
5. the updated dependency direction and this specification.

Until that design exists, database-specific code, dependencies, and migration
files are outside the established project architecture.
