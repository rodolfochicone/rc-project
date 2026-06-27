# Code Review Checklist

Language- and framework-agnostic review criteria. Detect the project's stack first,
then translate each universal concern into that stack's idioms, tooling, and
conventions. Honor conventions documented in the project (CLAUDE.md, lint config,
ADRs) over generic preference.

## Severity Levels

- **critical** — Security flaws, crashes, data loss, undefined behavior, or race
  conditions that could cause incidents or compromise data. E.g. auth bypass,
  injection, nil dereference on a hot path, unbounded resource leak, secrets in logs.
- **high** — Correctness bugs, user-visible performance bottlenecks, or anti-patterns
  that significantly impair scalability or reliability. Fix before merge. E.g. logic
  error returning wrong results, O(n²) over unbounded input, missing rollback,
  silently swallowed error on a critical path, missing boundary validation.
- **medium** — Maintainability concerns, code smells, or coverage gaps. Not blocking
  but should be addressed. E.g. duplicated logic, an over-long function with deep
  nesting, a missing test for an error branch, a single-use abstraction.
- **low** — Minor improvements: naming, docs, redundant conversions, misleading comments.

## Evaluation Areas

### 1. Security (ground in the OWASP Top 10)
- Injection — SQL, command, LDAP, template: parameterized queries / prepared statements only.
- Broken access control and authorization: privilege escalation, missing ownership checks.
- Authentication and session handling weaknesses.
- Cross-site scripting and missing output encoding.
- Hardcoded secrets, tokens, or credentials; weak or misused cryptography.
- Input validation at every trust boundary.
- Sensitive data exposure in logs, errors, or responses (no stack traces to clients).
- Context/identity headers read from a trusted source, never from the external client.

### 2. Correctness
- Logic errors and wrong results; off-by-one and boundary bugs.
- Null/nil/undefined dereferences; unsafe casts or coercions.
- Unhandled error or rejection paths leading to silent failure.

### 3. Concurrency & Async
- Races and missing synchronization on shared state.
- Leaked threads/goroutines/tasks/promises with no cancellation or join path.
- Deadlock potential from lock or resource ordering; misuse of concurrency primitives.
- Background work spawned without ownership or shutdown tracking.

### 4. Performance & Scalability
- Algorithmic complexity (O(n²) where O(n) suffices); N+1 queries.
- Resource leaks (file handles, sockets, response bodies, DB connections).
- Unbounded growth in collections, buffers, or queues; missing pagination.
- Missing caching for repeated expensive work; blocking I/O without timeout on hot paths.

### 5. Error Handling
- Swallowed/ignored errors without justification; lost error context when propagating.
- Crashing (panic/fatal/uncaught) in library or handler code instead of a recoverable error.
- Broad catch-all that masks specific failures; inconsistent error matching.

### 6. Code Quality & Maintainability
- Unclear naming, deeply nested logic, over-complex functions that should be decomposed.
- Duplication across functions or modules; dead code or unused exports.
- Comments that explain "what" instead of "why"; missing rationale for non-obvious choices.

### 7. Testing
- Missing tests for critical paths and error branches.
- Tests that verify mocks instead of behavior, or that cannot fail when business logic changes.
- Flaky patterns (time- or order-dependent); missing isolation/determinism.

### 8. Architecture
- Circular dependencies; layer violations (presentation reaching into runtime internals).
- Leaky abstractions; tight coupling that blocks independent testing.
- Patterns inconsistent with the surrounding codebase area.

### 9. Project-Convention Conformance
- Naming, file layout, dependency-injection, and logging conventions documented in the repo.
- Public API/response contracts and HTTP status-code conventions (never 200 with an error body).
- New dependencies wired wherever the project requires (DI registries, config, both client/server paths).

## Review Approach (after Google eng-practices)
- The goal is the codebase's long-term health, not perfection — approve a clear improvement.
- Review every line in scope; assess design, functionality, complexity, tests, naming, comments.
- Validate behavior and edge cases, not just shape; confirm the change does what the author intended.
- Reward simplicity; flag speculative or over-engineered complexity.
- Provide actionable feedback: state the problem and what the fix looks like.
- Assign severity by impact, not theoretical concern; acknowledge good work.

## Feedback Format (Conventional Comments)
Use `<label> [decoration]: <subject>` so severity and expectation are machine-clear:
- Labels: `issue`, `suggestion`, `nitpick`, `question`, `praise`, `todo`, `note`.
- Decorations: `(blocking)` must resolve before merge; `(non-blocking)` does not block;
  `(if-minor)` at the author's discretion. Tag the area too, e.g. `[security]`, `[test]`.
