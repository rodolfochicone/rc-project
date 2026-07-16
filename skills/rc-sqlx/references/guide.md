# SQLx 0.8 + Postgres — guide

Sources: docs.rs/sqlx (latest 0.8 line).

## Cargo features

```toml
sqlx = { version = "0.8", features = [
  "runtime-tokio",
  "postgres",
  "chrono",
  "uuid",
  "migrate",   # if using Migrator
] }
```

TLS: add `rustls` or `tls-rustls` feature set as required by the deployment environment.

## Pool

```rust
use sqlx::postgres::PgPoolOptions;
use sqlx::PgPool;

pub async fn connect(database_url: &str) -> Result<PgPool, sqlx::Error> {
    PgPoolOptions::new()
        .max_connections(10)
        .acquire_timeout(std::time::Duration::from_secs(5))
        .connect(database_url)
        .await
}
```

- One pool per process; clone `PgPool` into Axum state (cheap).
- Size ≈ concurrent in-flight queries; leave headroom for admin/migrations.

## Runtime queries (no compile-time DB)

```rust
use sqlx::FromRow;
use chrono::{DateTime, Utc};

#[derive(Debug, FromRow)]
struct Metric {
    id: i64,
    name: String,
    value: f64,
    recorded_at: DateTime<Utc>,
}

async fn list_metrics(pool: &PgPool) -> Result<Vec<Metric>, sqlx::Error> {
    sqlx::query_as::<_, Metric>(
        r#"
        SELECT id, name, value, recorded_at
        FROM metrics
        ORDER BY recorded_at DESC
        LIMIT 50
        "#,
    )
    .fetch_all(pool)
    .await
}

async fn insert_metric(pool: &PgPool, name: &str, value: f64) -> Result<Metric, sqlx::Error> {
    sqlx::query_as::<_, Metric>(
        r#"
        INSERT INTO metrics (name, value)
        VALUES ($1, $2)
        RETURNING id, name, value, recorded_at
        "#,
    )
    .bind(name)
    .bind(value)
    .fetch_one(pool)
    .await
}
```

## Bind parameters (required)

Postgres placeholders: `$1`, `$2`, … (reusable).

```rust
// CORRECT
sqlx::query("SELECT * FROM users WHERE email = $1")
    .bind(email)
    .fetch_optional(pool)
    .await?;

// WRONG — injection risk
// let q = format!("SELECT * FROM users WHERE email = '{}'", email);
```

## Compile-time macros

When `DATABASE_URL` is set at compile time (or offline data present):

```rust
let account = sqlx::query_as!(
    Account,
    r#"SELECT id, name FROM accounts WHERE id = $1"#,
    id
)
.fetch_one(pool)
.await?;
```

Type override syntax:

```rust
sqlx::query_as!(
    MyUser,
    r#"SELECT id as "id: UserId", name FROM users WHERE id = $1"#,
    id
)
```

Prepare offline:

```bash
cargo sqlx prepare --workspace
# CI: SQLX_OFFLINE=true cargo build
```

## Transactions

```rust
let mut tx = pool.begin().await?;

sqlx::query("UPDATE accounts SET balance = balance - $1 WHERE id = $2")
    .bind(amount)
    .bind(from_id)
    .execute(&mut *tx)
    .await?;

sqlx::query("UPDATE accounts SET balance = balance + $1 WHERE id = $2")
    .bind(amount)
    .bind(to_id)
    .execute(&mut *tx)
    .await?;

tx.commit().await?;
```

Drop without commit → rollback.

## Fetch modes

| Method | Use |
| ------ | --- |
| `fetch_one` | exactly one row required |
| `fetch_optional` | 0 or 1 |
| `fetch_all` | list (watch unbounded selects) |
| `fetch` | stream rows |

Always `LIMIT` list endpoints unless intentional full scan.

## JSON / chrono / uuid

Enable matching sqlx features; use `Json<T>`, `DateTime<Utc>`, `Uuid` in structs with `FromRow`.

## Layering with Axum

```rust
#[derive(Clone)]
struct AppState {
    pool: PgPool,
}

async fn handler(State(state): State<AppState>) -> Result<Json<Vec<Metric>>, AppError> {
    let rows = list_metrics(&state.pool).await.map_err(AppError::from)?;
    Ok(Json(rows))
}
```

Keep SQL in a `db/` or `repo/` module — not inline spaghetti in every handler.
