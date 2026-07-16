use crate::state::AppState;
use axum::extract::ws::{Message, WebSocket};
use axum::extract::{State, WebSocketUpgrade};
use axum::response::IntoResponse;
use axum::routing::get;
use axum::{Json, Router};
use chrono::{DateTime, Utc};
use futures_util::{SinkExt, StreamExt};
use serde::Serialize;
use sqlx::FromRow;
use std::time::Duration;
use tokio::time::interval;

pub fn router() -> Router<AppState> {
    Router::new()
        .route("/health", get(health))
        .route("/api/hello", get(hello))
        .route("/api/metrics", get(list_metrics).post(insert_metric))
        .route("/ws/metrics", get(ws_metrics))
}

async fn health() -> impl IntoResponse {
    Json(serde_json::json!({ "ok": true }))
}

async fn hello() -> impl IntoResponse {
    Json(serde_json::json!({
        "message": "Hello from Axum",
        "ts": Utc::now().to_rfc3339()
    }))
}

#[derive(Debug, Serialize, FromRow)]
struct Metric {
    id: i64,
    name: String,
    value: f64,
    recorded_at: DateTime<Utc>,
}

async fn list_metrics(State(state): State<AppState>) -> impl IntoResponse {
    let rows = sqlx::query_as::<_, Metric>(
        r#"
        SELECT id, name, value, recorded_at
        FROM metrics
        ORDER BY recorded_at DESC
        LIMIT 50
        "#,
    )
    .fetch_all(&state.pool)
    .await
    .unwrap_or_default();

    Json(rows)
}

#[derive(serde::Deserialize)]
struct NewMetric {
    name: String,
    value: f64,
}

async fn insert_metric(
    State(state): State<AppState>,
    Json(body): Json<NewMetric>,
) -> impl IntoResponse {
    let row = sqlx::query_as::<_, Metric>(
        r#"
        INSERT INTO metrics (name, value)
        VALUES ($1, $2)
        RETURNING id, name, value, recorded_at
        "#,
    )
    .bind(&body.name)
    .bind(body.value)
    .fetch_one(&state.pool)
    .await
    .expect("insert metric failed");

    Json(row)
}

async fn ws_metrics(ws: WebSocketUpgrade, State(state): State<AppState>) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_ws(socket, state))
}

async fn handle_ws(socket: WebSocket, state: AppState) {
    let (mut sender, mut receiver) = socket.split();
    let mut tick = interval(Duration::from_secs(2));

    loop {
        tokio::select! {
            _ = tick.tick() => {
                let sample = MetricSample {
                    name: "cpu_mock".into(),
                    value: (Utc::now().timestamp_subsec_millis() as f64) / 10.0 % 100.0,
                    recorded_at: Utc::now(),
                };

                let _ = sqlx::query(
                    "INSERT INTO metrics (name, value) VALUES ($1, $2)",
                )
                .bind(&sample.name)
                .bind(sample.value)
                .execute(&state.pool)
                .await;

                let payload = serde_json::to_string(&sample).unwrap_or_default();
                if sender.send(Message::Text(payload.into())).await.is_err() {
                    break;
                }
            }
            msg = receiver.next() => {
                match msg {
                    Some(Ok(Message::Close(_))) | None => break,
                    Some(Ok(Message::Ping(p))) => {
                        if sender.send(Message::Pong(p)).await.is_err() {
                            break;
                        }
                    }
                    _ => {}
                }
            }
        }
    }
}

#[derive(Serialize)]
struct MetricSample {
    name: String,
    value: f64,
    recorded_at: DateTime<Utc>,
}
