# Review Criteria

These criteria are language- and framework-agnostic. Apply them to whatever
stack the reviewed project uses. Before reviewing, identify the project's stack
(see "Stack-Specific Best Practices" below) and translate each universal concern
into the idioms, tooling, and conventions of that stack.

## Severity Levels

### critical

Security flaws, crashes, data loss, undefined behavior, or race conditions.
Issues that could cause production incidents or compromise user data.

Examples: authentication bypass, SQL/command injection, null/nil dereference
on a hot path, unbounded resource or concurrency leak, writing sensitive data
to logs.

### high

Bugs affecting correctness, performance bottlenecks visible to users, or
anti-patterns that significantly impair scalability, reliability, or usability.
These need fixing before merge.

Examples: logic error returning wrong results, O(n^2) loop over unbounded
input, missing transaction rollback, error silently swallowed in a critical
path, missing input validation at a system boundary.

### medium

Maintainability concerns, code smells, test coverage gaps, or non-idiomatic
patterns that degrade long-term health. Not blocking but should be addressed.

Examples: duplicated logic across modules, function exceeding the project's
length/complexity norms with deep nesting, missing test for an error branch,
an abstraction introduced for a single implementation.

### low

Minor improvements, documentation gaps, or naming suggestions. Optional
enhancements that improve clarity.

Examples: unclear variable name, missing doc comment on a public API,
redundant type conversion, slightly misleading comment.

## Evaluation Areas

### 1. Security

- Authentication and authorization flaws.
- Input validation gaps (injection, path traversal, XSS).
- Hardcoded secrets, tokens, or credentials.
- Cryptography misuse or insecure storage.
- Sensitive data exposure in logs or error messages.

### 2. Correctness

- Logic errors producing wrong results.
- Off-by-one and boundary condition bugs.
- Null, nil, or undefined dereferences.
- Unhandled error or rejection paths leading to silent failures.
- Incorrect type coercions, casts, or conversions.

### 3. Concurrency and Async

- Race conditions and missing synchronization on shared state.
- Leaked threads, tasks, coroutines, or promises (no cancellation or join path).
- Deadlock potential from lock or resource ordering.
- Misuse of the stack's concurrency primitives (channels, locks, async/await,
  event loops, worker pools).
- Spawned background work without ownership or shutdown tracking.

### 4. Performance and Scalability

- Algorithmic complexity issues (O(n^2) where O(n) suffices).
- Resource leaks (file handles, sockets, response bodies, database connections).
- Unbounded growth in collections, buffers, or queues.
- Missing caching for repeated expensive operations.
- Blocking I/O on critical paths without timeout.

### 5. Error Handling

- Swallowed or ignored errors and rejections without justification.
- Missing error context when propagating failures up the stack.
- Crashing the process (uncaught exceptions, fatal aborts) in library or
  handler code instead of returning a recoverable error.
- Broad catch-all handling that masks specific failures.
- Inconsistent or incorrect error-matching logic.

### 6. Code Quality and Maintainability

- Readability issues (unclear naming, deeply nested logic).
- Code duplication across functions or modules.
- Overly complex functions that should be decomposed.
- Dead code or unused exports.
- Violations of project coding conventions.

### 7. Testing

- Missing tests for critical code paths.
- Tests that verify mocks instead of behavior.
- Flaky test patterns (time-dependent, order-dependent).
- Inadequate edge case and error path coverage.
- Missing isolation or determinism the project's test conventions expect.

### 8. Architecture

- Circular dependencies between modules or packages.
- Layer violations (e.g., a presentation layer reaching into internal runtime
  details).
- Leaky abstractions exposing implementation details.
- Tight coupling that prevents independent testing.
- Inconsistent patterns within the same codebase area.

### 9. Operations

- Missing or insufficient structured logging.
- Missing error context for production debugging.
- Configuration values hardcoded instead of parameterized.
- Missing graceful shutdown handling for long-running processes.
- Observability gaps (no metrics or tracing on critical operations).

## Stack-Specific Best Practices

The universal areas above describe *what* to look for. *How* each concern
manifests depends on the stack. Before assigning severity, ground the review in
the project's actual technology:

- Detect the stack from manifest and config files (e.g., `go.mod`,
  `package.json`, `pyproject.toml`, `Cargo.toml`, `pom.xml`, `*.csproj`,
  `Gemfile`) and from the project's `CLAUDE.md` or contributing guide.
- Apply the idiomatic best practices, style guide, and anti-patterns of that
  language and its frameworks — error handling, concurrency model, logging,
  dependency injection, testing conventions, and naming all differ per stack.
- Honor conventions already documented in the project (CLAUDE.md, lint config,
  editor config, ADRs) over generic preferences. Conformance to the existing
  codebase outranks personal taste.
- When the correct idiom for an unfamiliar stack is unclear, consult the
  ecosystem's authoritative documentation before flagging it as an issue rather
  than guessing.
- Express each finding in the target stack's terms so the suggested fix is
  directly actionable in that language.

## Review Approach

- Read the PRD and TechSpec before reviewing code to understand intent.
- Review in severity order: critical first, low last.
- Focus on issues that matter. Skip style issues already caught by linters.
- Provide actionable suggestions: state the problem and what the fix looks like.
- Assign severity based on actual impact, not theoretical concern.
- Create one issue per file per distinct problem.
- If one problem spans multiple files, create one issue per affected file.
- Acknowledge well-implemented patterns; do not create issues for them.
