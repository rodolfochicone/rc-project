# Axum / Rust — lint & tooling

## Local gates (run before claim-done)

```bash
cargo fmt --all -- --check
cargo clippy --all-targets --all-features -- -D warnings
cargo test
cargo build --release
```

## rustfmt

- Commit `rustfmt.toml` only if the team needs non-defaults.
- Never mix hand-aligned style with unformatted modules.

## Clippy

Deny warnings in CI for the app crate:

```bash
cargo clippy -- -D warnings
```

Useful lints for web services:

- `unwrap_used` / `expect_used` (pedantic; allow in tests if needed)
- `panic` in non-test code
- large future / async mistakes

## Features & bloat

- Enable only needed `tower-http` features (`cors`, `trace`, `limit`, …).
- Prefer `axum` features `ws` only if used.

## Tracing

```rust
tracing_subscriber::fmt()
    .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
    .init();
```

Default filter example: `RUST_LOG=app_api=info,tower_http=info`.

## CI sketch

```yaml
# conceptual
- cargo fmt --check
- cargo clippy --all-targets -- -D warnings
- cargo test
- cargo audit  # optional, needs cargo-audit
```

## Edition & MSRV

- Edition **2021** (or 2024 when the workspace standardizes).
- Document MSRV if the library is published; binaries can track stable latest.

## Pair with

- `rc-sqlx` for DB-related clippy/sqlx offline mode
- `rc-final-verify` after green gates
