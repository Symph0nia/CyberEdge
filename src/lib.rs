use axum::{Json, Router, routing::get};
use serde::Serialize;

#[derive(Debug, PartialEq, Serialize)]
struct Health {
    status: &'static str,
}

pub fn app() -> Router {
    Router::new().route("/api/v1/health", get(health))
}

async fn health() -> Json<Health> {
    Json(Health { status: "ok" })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn health_is_ok() {
        assert_eq!(health().await.0, Health { status: "ok" });
    }
}
