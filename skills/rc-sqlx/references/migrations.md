# SQLx migrations

## Layout

```text
migrations/
  20260101120000_init.sql
  20260102140000_add_metrics.sql
```

Or timestamp-prefixed folders depending on sqlx-cli version conventions — stay consistent in the repo.

## sqlx-cli

```bash
cargo install sqlx-cli --no-default-features --features rustls,postgres
export DATABASE_URL=postgres://app:app@127.0.0.1:5432/app
sqlx database create
sqlx migrate run
sqlx migrate info
```

## Embedded migrator (app startup)

```rust
sqlx::migrate!("./migrations")
    .run(&pool)
    .await?;
```

Use carefully in multi-instance deploys: only one migrator should run (job, or advisory lock pattern). Prefer migrate-in-deploy **before** switching traffic when possible.

## Rules

1. **Forward-only in prod** when possible; document manual down steps for rollbacks.
2. Avoid destructive changes without a expand/contract plan (add column → backfill → switch → drop).
3. Keep migrations **idempotent enough** for re-runs only if sqlx tracks version table — do not hand-edit applied files; add a new migration.
4. Test migrations against a copy of prod-like data when changing large tables.

## Example migration

```sql
-- migrations/20260716000000_metrics.sql
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS metrics_recorded_at_idx ON metrics (recorded_at DESC);
```

## CI

- Run `sqlx migrate run` against ephemeral Postgres service.
- If using query macros: `cargo sqlx prepare --check` to fail on stale `.sqlx` offline data.
