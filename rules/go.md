---
paths: ["**/*.go", "**/go.mod", "**/go.sum"]
---

# Go rules

The WHAT for Go code. Enforced as patterns; some are also backed by the
`go-mod-guard` and `go-fmt` hooks and the project's verification gate.

- Wrap errors with `fmt.Errorf("context: %w", err)`; match with `errors.Is` / `errors.As`. Never compare error strings.
- Handle every error explicitly. No `_` discards without a written justification.
- No `panic()` / `log.Fatal()` in production paths — only truly unrecoverable startup failures.
- Use `log/slog` for operational logging, not `log.Printf` / `fmt.Println`.
- Pass `context.Context` as the first arg across runtime boundaries; avoid `context.Background()` outside `main` and focused tests.
- Small, focused interfaces: accept interfaces, return structs. Functional options for complex constructors. Compile-time checks via `var _ Iface = (*T)(nil)`.
- Avoid `interface{}`/`any` when a concrete type is known; avoid reflection without a performance justification.
- Every goroutine has explicit ownership and shutdown via `context` cancellation; track with `sync.WaitGroup`. No fire-and-forget. `select` on `ctx.Done()` in long-running loops.
- `sync.RWMutex` for read-heavy state, `sync.Mutex` for write-heavy; prefer channels over shared memory when practical.
- Tests are table-driven with `t.Run` subtests, `t.Parallel()` where independent, `t.TempDir()` for fs isolation, `t.Helper()` in helpers; must pass with `-race`. Mock via interfaces, not test-only methods in production code.
- Never edit `go.mod` / `go.sum` by hand — use `go get` / `go mod tidy`.
- Format with `gofmt`; `make lint` (golangci-lint) has zero tolerance.
