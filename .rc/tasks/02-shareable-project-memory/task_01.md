---
status: completed
title: "Mirror format — marshal, parse, and filename for memory files"
type: backend
complexity: medium
dependencies: []
---

# Task 1: Mirror format — marshal, parse, and filename for memory files

## Overview
Define the on-disk text representation of a project memory: a markdown file with YAML
frontmatter plus body. This task adds pure functions to serialize a `Memory` to that format,
parse it back losslessly, and derive a deterministic filename — the foundation both the
export and import subcommands build on.

<critical>
- ALWAYS READ the TechSpec ("Implementation Design" and "Behavioral Contract") and ADR-001 before starting
- REFERENCE TECHSPEC for implementation details — do not duplicate here
- FOCUS ON "WHAT" — describe what needs to be accomplished, not how
- MINIMIZE CODE — show code only to illustrate current structure or problem areas
- TESTS REQUIRED — every task MUST include tests in deliverables
</critical>

<requirements>
- MUST add `MarshalMemory(m Memory) ([]byte, error)`, `ParseMemory(data []byte) (Memory, error)`, and `MirrorFileName(m Memory) string` to package `projectmemory` in a new file.
- MUST serialize frontmatter fields `id, scope, key, title, tags, source, created_at, updated_at` and the body as file content, reusing `internal/core/frontmatter` (`Format`/`Parse`) rather than hand-rolling a parser.
- MUST satisfy `ParseMemory(MarshalMemory(m)) == m` for all fields including tag set and UTC timestamps (contract `memory-mirror.roundtrip-lossless`).
- MUST name keyed memories `<scope>__<sanitized-key>.md` and keyless memories `<id>.md`, where sanitization lowercases and replaces every non-alphanumeric run with `-` (contracts `memory-mirror.keyed-stable-filename`, `memory-mirror.keyless-id-filename`).
- MUST return a descriptive error from `ParseMemory` when a required field (id, scope, title, body, created_at, updated_at) is missing or timestamps are unparseable.
- MUST format/parse timestamps with `store.FormatTimestamp`/`store.ParseTimestamp` to match the rest of the store.
</requirements>

## Subtasks
- [x] 1.1 Add a new file in `internal/core/projectmemory` for the mirror format (`mirror.go`).
- [x] 1.2 Implement `MarshalMemory` using `frontmatter.Format` with a private frontmatter struct.
- [x] 1.3 Implement `ParseMemory` using `frontmatter.Parse`, validating required fields and timestamps.
- [x] 1.4 Implement `MirrorFileName` with deterministic, idempotent key sanitization.
- [x] 1.5 Add table-driven unit tests covering round-trip, filename rules, and parse errors.

## Implementation Details
Add the three functions to a new file under `internal/core/projectmemory`. See TechSpec
"Core Interfaces" for signatures and "Data Models" for the file shape. The frontmatter struct
should carry the eight fields as YAML tags; reuse `frontmatter.Format`/`frontmatter.Parse`
(the same package `internal/core/tasks/parser.go` uses) so the format matches existing
artifacts. Tags should round-trip through the store's normalized form (`splitTags` produces
the slice; serialize as a YAML list).

### Relevant Files
- `internal/core/projectmemory/store.go` — `Memory` struct (lines 38-49) and the `splitTags`/`normalizeTags` helpers this code reuses.
- `internal/core/frontmatter/frontmatter.go` — `Format[T]`/`Parse[T]` to serialize and parse frontmatter.
- `internal/core/tasks/parser.go` — precedent for `frontmatter.Parse` into a typed struct.
- `internal/store` — `FormatTimestamp`/`ParseTimestamp` for UTC timestamp handling.

### Dependent Files
- `internal/cli/memory.go` — task_03/task_04 call these functions from the new subcommands.

### Patterns to Mirror
```go
// SOURCE: internal/core/tasks/parser.go:49-55
func ParseTaskFile(content string) (model.TaskEntry, error) {
	var node yaml.Node
	if _, err := frontmatter.Parse(content, &node); err != nil {
		...
		return model.TaskEntry{}, fmt.Errorf("parse task front matter: %w", err)
	}
```

```go
// SOURCE: internal/core/projectmemory/store.go:345-358
func hydrateMemory(memory Memory, key sql.NullString, tags, created, updated string) (Memory, error) {
	createdAt, err := store.ParseTimestamp(created)
	if err != nil {
		return Memory{}, fmt.Errorf("projectmemory: parse created_at: %w", err)
	}
	...
}
```

### Related ADRs
- [ADR-001: Share project memory through a committed markdown-per-fact mirror](adrs/adr-001.md) — defines the markdown-per-fact format.
- [ADR-002: Mirror identity and most-recent-wins conflict resolution](adrs/adr-002.md) — defines the filename identity rule.

## Deliverables
- New `internal/core/projectmemory/mirror.go` with `MarshalMemory`, `ParseMemory`, `MirrorFileName`.
- Unit tests with 80%+ coverage **(REQUIRED)**.
- Round-trip and filename tests proving the behavioral contract ids above **(REQUIRED)**.

## Tests
- Unit tests:
  - [ ] Round-trip: a memory with key, tags, and source survives `Marshal`→`Parse` unchanged (all fields, tag order, UTC timestamps).
  - [ ] Round-trip: a keyless memory with no tags and empty source survives unchanged.
  - [ ] Filename: keyed memory `scope=decision, key=project-memory-sharing` → `decision__project-memory-sharing.md`.
  - [ ] Filename: key with spaces/uppercase/slashes sanitizes deterministically and is idempotent on re-marshal.
  - [ ] Filename: keyless memory `id=mem-abc123` → `mem-abc123.md`.
  - [ ] Parse error: file missing `title` returns an error naming the field.
  - [ ] Parse error: unparseable `updated_at` returns an error.
- Integration tests:
  - [ ] Not applicable for this task (pure functions); end-to-end coverage lands in task_04.
- Test coverage target: >=80%
- All tests must pass

## Success Criteria
- All tests passing
- Test coverage >=80%
- `ParseMemory(MarshalMemory(m))` is lossless for every field
- Filenames are deterministic and stable across machines for the same identity
