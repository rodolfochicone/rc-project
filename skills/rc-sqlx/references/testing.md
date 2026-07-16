# SQLx — testing

## Principles

- Prefer a **real Postgres** (Docker) over mocking the pool for repository code.
- Isolate tests: transaction rollback, unique DB per test, or `sqlx::test` database.

## `sqlx::test` (integration)

With the `sqlx` test macro (see crate docs for exact attributes on your version):

```rust
#[sqlx::test]
async fn insert_and_list(pool: PgPool) -> sqlx::Result<()> {
    sqlx::query("INSERT INTO metrics (name, value) VALUES ($1, $2)")
        .bind("cpu")
        .bind(1.0_f64)
        .execute(&pool)
        .await?;

    let rows = sqlx::query_as::<_, Metric>("SELECT id, name, value, recorded_at FROM metrics")
        .fetch_all(&pool)
        .await?;

    assert_eq!(rows.len(), 1);
    assert_eq!(rows[0].name, "cpu");
    Ok(())
}
```

Ensure migrations apply so schema exists (migrator attribute / automatic migrate per sqlx version).

## Manual docker

```bash
docker run -d --name pg-test -e POSTGRES_PASSWORD=test -p 5433:5432 postgres:16-alpine
export DATABASE_URL=postgres://postgres:test@127.0.0.1:5433/postgres
sqlx migrate run
cargo test
```

## Unit tests without DB

- Pure functions (mapping, allowlists, DTO validation) stay DB-free.
- Do not invent a fake SQL engine unless necessary.

## Assertions that matter

- Constraint violations surface as expected errors
- Transactions roll back on failure
- Idempotent upserts behave correctly
- Pagination boundaries

## Commands

```bash
cargo test
SQLX_OFFLINE=true cargo test   # if offline mode is the project default
```

## Anti-patterns

- Sharing mutable global pool state across parallel tests without isolation
- Pointing tests at production
- Asserting only `is_ok()` without row content
