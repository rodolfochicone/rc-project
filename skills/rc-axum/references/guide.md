# Axum 0.8 — guide

Sources: Axum docs (tokio-rs/axum v0.8.x), tower-http patterns.

## Minimal server

```rust
use axum::{Router, routing::get};
use std::net::SocketAddr;

#[tokio::main]
async fn main() {
    let app = Router::new().route("/health", get(|| async { "ok" }));
    let addr = SocketAddr::from(([127, 0, 0, 1], 3000));
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
```

Prefer **127.0.0.1** behind Caddy/Nginx on a VPS; let the proxy terminate TLS.

## App state

```rust
use axum::{extract::State, routing::get, Router};

#[derive(Clone)]
struct AppState {
    // PgPool is Clone; wrap non-Clone clients in Arc
}

async fn handler(State(state): State<AppState>) -> &'static str {
    let _ = state;
    "ok"
}

let app = Router::new()
    .route("/", get(handler))
    .with_state(AppState {});
```

- **App-wide:** `State<AppState>` + `.with_state(...)`.
- **Request-scoped** (e.g. `CurrentUser` after auth middleware): `Extension<T>` or a custom extractor.

## Extractors (order & failure)

Common extractors: `Path`, `Query`, `Json`, `State`, `Headers` / `TypedHeader`, `Multipart`.

If an extractor fails, Axum rejects the request before the handler runs. Map rejection types to clear 4xx bodies via custom extractors or `FromRequest` when the default is too opaque.

```rust
use axum::{extract::Path, Json};
use serde::Deserialize;

#[derive(Deserialize)]
struct CreateBody {
    name: String,
}

async fn create(
    Path(id): Path<uuid::Uuid>,
    Json(body): Json<CreateBody>,
) -> impl axum::response::IntoResponse {
    let _ = (id, body);
    axum::http::StatusCode::CREATED
}
```

## Nested routers & modular layout

```text
src/
  main.rs          # serve + tracing init
  state.rs
  error.rs         # AppError: IntoResponse
  routes/
    mod.rs         # Router::merge
    health.rs
    metrics.rs
  middleware/
    auth.rs
```

```rust
fn api_router() -> Router<AppState> {
    Router::new()
        .route("/health", get(health))
        .nest("/api", resource_router())
}
```

## Middleware (Tower)

Order matters (outer first):

1. `HandleErrorLayer` + timeout
2. `TraceLayer`
3. `CorsLayer` (explicit origins in prod)
4. Compression / request body limits
5. Auth as **`route_layer`** on protected routes only

```rust
use axum::{
    error_handling::HandleErrorLayer,
    BoxError,
    http::{Method, StatusCode, Uri},
};
use std::time::Duration;
use tower::ServiceBuilder;
use tower_http::{cors::CorsLayer, trace::TraceLayer};

async fn handle_timeout_error(method: Method, uri: Uri, err: BoxError) -> (StatusCode, String) {
    (
        StatusCode::GATEWAY_TIMEOUT,
        format!("{method} {uri} failed: {err}"),
    )
}

let app = Router::new()
    .merge(api_router())
    .layer(
        ServiceBuilder::new()
            .layer(HandleErrorLayer::new(handle_timeout_error))
            .timeout(Duration::from_secs(30))
            .layer(TraceLayer::new_for_http())
            .layer(CorsLayer::new()), // configure allow_origin explicitly
    )
    .with_state(state);
```

### Auth middleware → handler

Insert user into extensions; extract with `Extension<CurrentUser>` (or a typed extractor).

```rust
use axum::{
    extract::{Request, Extension},
    middleware::{self, Next},
    response::Response,
    http::StatusCode,
};

#[derive(Clone)]
struct CurrentUser { id: uuid::Uuid }

async fn auth(mut req: Request, next: Next) -> Result<Response, StatusCode> {
    let token = req
        .headers()
        .get(axum::http::header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .ok_or(StatusCode::UNAUTHORIZED)?;
    let user = authorize(token).await.ok_or(StatusCode::UNAUTHORIZED)?;
    req.extensions_mut().insert(user);
    Ok(next.run(req).await)
}

// .route_layer(middleware::from_fn(auth)) on protected routes
```

## Typed errors

```rust
use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;

pub enum AppError {
    NotFound,
    BadRequest(String),
    Internal(anyhow::Error),
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, msg) = match &self {
            AppError::NotFound => (StatusCode::NOT_FOUND, "not found".into()),
            AppError::BadRequest(m) => (StatusCode::BAD_REQUEST, m.clone()),
            AppError::Internal(e) => {
                tracing::error!(error = %e, "internal");
                (StatusCode::INTERNAL_SERVER_ERROR, "internal error".into())
            }
        };
        (status, Json(json!({ "error": msg }))).into_response()
    }
}

type AppResult<T> = Result<T, AppError>;
```

## WebSockets

Enable feature `ws` on axum. Upgrade in the handler; do auth **before** `on_upgrade` when the stream is private.

```rust
use axum::{
    extract::ws::{Message, WebSocket, WebSocketUpgrade},
    response::IntoResponse,
};
use futures_util::{SinkExt, StreamExt};

async fn ws_handler(ws: WebSocketUpgrade) -> impl IntoResponse {
    ws.on_upgrade(handle_socket)
}

async fn handle_socket(mut socket: WebSocket) {
    while let Some(Ok(msg)) = socket.recv().await {
        match msg {
            Message::Text(t) => {
                if socket.send(Message::Text(t)).await.is_err() {
                    break;
                }
            }
            Message::Close(_) => break,
            // Ping is answered by the stack; handle only if logging is needed
            _ => {}
        }
    }
}
```

Production WS checklist:

- Cap frame size / disconnect slow clients
- Heartbeat or idle timeout
- Fan-out via channels/`broadcast` — do not block on slow subscribers forever
- Same-origin or explicit CORS/WS origin checks at proxy

## JSON responses

Prefer `Json<T>` with serde types. For empty success use `StatusCode`. Avoid returning raw `String` for API payloads.

## Feature flags (Cargo.toml sketch)

```toml
axum = { version = "0.8", features = ["ws", "macros"] }
tokio = { version = "1", features = ["full"] }
tower = { version = "0.5", features = ["timeout"] }
tower-http = { version = "0.6", features = ["cors", "trace", "limit"] }
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
```

Pin compatible versions in the app lockfile; bump intentionally.
