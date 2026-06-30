# Simplify Review — 02-shareable-project-memory

Scope: the uncommitted change set for shareable project memory —
`internal/core/projectmemory/{mirror.go, mirror_test.go, import.go, import_test.go}`,
`internal/cli/{memory_sync.go, memory_sync_test.go, memory_sync_import_test.go}`, and edits to
`internal/cli/memory.go`, `internal/core/projectmemory/store.go`, `skills/rc-project-memory/SKILL.md`.

Lens: over-engineering / complexity only. Correctness, security, and performance are out of
scope (routed to `rc-code-review` / `rc-review-round`).

## Findings

No provable cuts.

Verification performed:

- `make verify` passes with golangci-lint's `unused` linter enabled, so there is no dead
  private code in the change set — an unused package-level function would fail the gate.
- Every exported symbol has a real caller: `MarshalMemory`/`MirrorFileName` are used by
  `rc memory export`; `ParseMemory` by `rc memory import`; `Store.Import`/`ImportResult` by the
  import command (and its tests).
- The implementation climbed the laziness ladder at write time: it reuses
  `internal/core/frontmatter` (`Format`/`Parse`) instead of a hand-rolled parser, the existing
  `Store.List` for export reads, and the `schema.go` transaction-with-rollback pattern for
  `Import`. No stdlib duplication, no new dependency.

Candidates considered and deliberately not flagged:

- `storedMemory` (`import.go`) — an 8-parameter helper, but it is called from both the insert
  and update paths to build the record handed to `maybeEmbed`; it removes duplication rather
  than adding indirection.
- `imported []Memory` accumulation + post-commit embed loop (`import.go`) — required, not
  speculative: embeddings are best-effort and must run *after* `tx.Commit()`, mirroring how
  `Add` embeds only once the row is durable. Embedding inside the transaction would be wrong.
- `skippedMirrorFile{path, err}` (`memory_sync.go`) — captures what the task explicitly
  requires (report each malformed file via `slog.Warn`); reducing it to a bare count would drop
  a requested behavior.
- `importOutcome` enum — three named outcomes read clearly and drive the result counters; no
  cheaper form is also as clear.

## Verdict

Lean already. Ship.

net: -0 lines, -0 deps possible.
