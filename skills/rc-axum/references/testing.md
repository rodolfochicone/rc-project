# Axum — testing

## Unit-style HTTP tests (`oneshot`)

```rust
use axum::{
    body::Body,
    http::{Request, StatusCode},
    routing::get,
    Router,
};
use tower::ServiceExt; // for `oneshot`

async fn health() -> &'static str { "ok" }

#[tokio::test]
async fn health_ok() {
    let app = Router::new().route("/health", get(health));

    let res = app
        .oneshot(
            Request::builder()
                .uri("/health")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();

    assert_eq!(res.status(), StatusCode::OK);
}
```

With state:

```rust
let app = Router::new()
    .route("/api/hello", get(hello))
    .with_state(test_state());
```

## JSON POST

```rust
use serde_json::json;

let res = app
    .oneshot(
        Request::builder()
            .method("POST")
            .uri("/api/metrics")
            .header("content-type", "application/json")
            .body(Body::from(serde_json::to_vec(&json!({
                "name": "cpu",
                "value": 1.0
            })).unwrap()))
            .unwrap(),
    )
    .await
    .unwrap();
assert_eq!(res.status(), StatusCode::OK);
```

## Shared test helper

```rust
fn test_app(state: AppState) -> Router {
    routes::router().with_state(state)
}
```

## Database

Prefer real Postgres in integration tests via `rc-sqlx` patterns (`sqlx::test` or testcontainers). Keep pure handler tests free of DB when logic is pure.

## WebSocket tests

Use axum’s WS test utilities or a client that performs the upgrade handshake. Assert:

- unauthorized upgrade → 401
- first message schema
- close handling

## What to assert

| Layer | Assert |
| ----- | ------ |
| Happy path | status + JSON shape |
| Validation | 400/422 body |
| Auth | 401/403 without token |
| Not found | 404 |
| Timeouts | optional; harder in unit tests |

## Commands

```bash
cargo test
cargo test -p app-api -- --nocapture
RUST_LOG=debug cargo test
```

## Anti-patterns

- Only testing “compiles”
- Hitting production DATABASE_URL in tests
- Ignoring status codes and only checking `is_ok()` on the service call
