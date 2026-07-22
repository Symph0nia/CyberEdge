use std::{net::SocketAddr, path::PathBuf, sync::Arc};

use axum::{
    Json, Router,
    extract::{Path, State},
    http::{HeaderName, HeaderValue, StatusCode},
    routing::{any, get},
};
use base64::{Engine, engine::general_purpose::STANDARD};
use serde_json::{Value, json};
use tower_http::{
    services::{ServeDir, ServeFile},
    set_header::SetResponseHeaderLayer,
};

use crate::{
    proto::{
        Asset, AssetChange, Certificate, ExposureChange, Schedule, Scope, Service, Task, Website,
    },
    repository::Repository,
};

#[derive(Clone)]
struct WebState {
    repository: Arc<dyn Repository>,
}

pub async fn serve_read_only_web(
    repository: Arc<dyn Repository>,
    address: SocketAddr,
    dist: PathBuf,
) -> std::io::Result<()> {
    let app = read_only_router(repository, dist);
    let listener = tokio::net::TcpListener::bind(address).await?;
    axum::serve(listener, app).await
}

pub fn read_only_router(repository: Arc<dyn Repository>, dist: PathBuf) -> Router {
    let index = dist.join("index.html");
    Router::new()
        .route("/api/v1/overview", get(overview))
        .route("/api/v1/scopes/{scope_id}/assets", get(assets))
        .route("/api/v1/scopes/{scope_id}/services", get(services))
        .route("/api/v1/scopes/{scope_id}/certificates", get(certificates))
        .route("/api/v1/scopes/{scope_id}/websites", get(websites))
        .route("/api/v1/schedules/{schedule_id}/exposure-changes", get(exposure_changes))
        .route("/api/v1/tasks/{task_id}/observations", get(observations))
        .route("/api/v1/evidence/{evidence_id}", get(evidence))
        .route("/api/{*path}", any(api_not_found))
        .fallback_service(ServeDir::new(dist).not_found_service(ServeFile::new(index)))
        .layer(SetResponseHeaderLayer::overriding(
            HeaderName::from_static("content-security-policy"),
            HeaderValue::from_static("default-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'"),
        ))
        .layer(SetResponseHeaderLayer::overriding(
            HeaderName::from_static("x-content-type-options"),
            HeaderValue::from_static("nosniff"),
        ))
        .layer(SetResponseHeaderLayer::overriding(
            HeaderName::from_static("referrer-policy"),
            HeaderValue::from_static("no-referrer"),
        ))
        .with_state(WebState { repository })
}

async fn api_not_found() -> (StatusCode, Json<Value>) {
    (
        StatusCode::NOT_FOUND,
        Json(json!({"error": "read-only API route not found"})),
    )
}

async fn overview(State(state): State<WebState>) -> Result<Json<Value>, WebError> {
    let model = state.repository.read_overview().await.map_err(WebError)?;
    Ok(Json(json!({
        "counts": {"scopes": model.scope_count, "tasks": model.task_count,
            "assets": model.asset_count, "schedules": model.schedule_count,
            "asset_changes": model.asset_change_count,
            "exposure_changes": model.exposure_change_count,
            "services": model.service_count,
            "certificates": model.certificate_count,
            "websites": model.website_count,
            "observations": model.observation_count,
            "evidence": model.evidence_count,
            "notifications_pending": model.notification_pending_count,
            "notifications_delivered": model.notification_delivered_count,
            "notifications_dead_letter": model.notification_dead_letter_count},
        "scopes": model.scopes.into_iter().map(scope_json).collect::<Vec<_>>(),
        "tasks": model.tasks.into_iter().map(task_json).collect::<Vec<_>>(),
        "assets": model.assets.into_iter().map(asset_json).collect::<Vec<_>>(),
        "schedules": model.schedules.into_iter().map(schedule_json).collect::<Vec<_>>(),
        "asset_changes": model.asset_changes.into_iter().map(asset_change_json).collect::<Vec<_>>(),
        "exposure_changes": model.exposure_changes.into_iter().map(exposure_change_json).collect::<Vec<_>>(),
        "services": model.services.into_iter().map(service_json).collect::<Vec<_>>(),
        "certificates": model.certificates.into_iter().map(certificate_json).collect::<Vec<_>>(),
        "websites": model.websites.into_iter().map(website_json).collect::<Vec<_>>(),
        "audit_events": model.audit_events.into_iter().map(|event| json!({
            "id": event.id, "request_id": event.request_id, "operation": event.operation,
            "agent_id": event.agent_id, "skill_name": event.skill_name,
            "skill_version": event.skill_version, "resource_kind": event.resource_kind,
            "resource_id": event.resource_id, "occurred_at": timestamp(event.occurred_at)
        })).collect::<Vec<_>>()
    })))
}

async fn assets(
    State(state): State<WebState>,
    Path(scope_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_assets(&scope_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"assets": values.into_iter().map(asset_json).collect::<Vec<_>>() }),
    ))
}

async fn observations(
    State(state): State<WebState>,
    Path(task_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_observations(&task_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"observations": values.into_iter().map(|item| json!({
        "id": item.id, "task_id": item.task_id, "asset_id": item.asset_id,
        "type": item.observation_type,
        "value": serde_json::from_str::<Value>(&item.value_json).unwrap_or(Value::String(item.value_json)),
        "evidence_id": item.evidence_id, "observed_at": timestamp(item.observed_at)
    })).collect::<Vec<_>>() }),
    ))
}

async fn services(
    State(state): State<WebState>,
    Path(scope_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_services(&scope_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"services": values.into_iter().map(service_json).collect::<Vec<_>>() }),
    ))
}

async fn certificates(
    State(state): State<WebState>,
    Path(scope_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_certificates(&scope_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"certificates": values.into_iter().map(certificate_json).collect::<Vec<_>>() }),
    ))
}

async fn websites(
    State(state): State<WebState>,
    Path(scope_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_websites(&scope_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"websites": values.into_iter().map(website_json).collect::<Vec<_>>() }),
    ))
}

async fn exposure_changes(
    State(state): State<WebState>,
    Path(schedule_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_exposure_changes(&schedule_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"changes": values.into_iter().map(exposure_change_json).collect::<Vec<_>>() }),
    ))
}

async fn evidence(
    State(state): State<WebState>,
    Path(evidence_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let item = state
        .repository
        .get_evidence(&evidence_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"id": item.id, "media_type": item.media_type, "sha256": item.sha256,
        "content_base64": STANDARD.encode(item.content), "created_at": timestamp(item.created_at)}),
    ))
}

fn scope_json(scope: Scope) -> Value {
    json!({"id": scope.id, "name": scope.name, "authorization_ref": scope.authorization_ref,
        "targets": scope.targets.into_iter().map(|target| json!({"kind": target.kind, "value": target.value})).collect::<Vec<_>>(),
        "created_at": timestamp(scope.created_at)})
}

fn task_json(task: Task) -> Value {
    json!({"id": task.id, "scope_id": task.scope_id, "policy_id": task.policy_id,
        "state": task.state, "created_at": timestamp(task.created_at),
        "updated_at": timestamp(task.updated_at), "schedule_id": task.schedule_id})
}

fn asset_json(asset: Asset) -> Value {
    json!({"id": asset.id, "scope_id": asset.scope_id, "kind": asset.kind, "value": asset.value,
        "first_seen_at": timestamp(asset.first_seen_at), "last_seen_at": timestamp(asset.last_seen_at)})
}

fn service_json(service: Service) -> Value {
    json!({"id": service.id, "asset_id": service.asset_id, "transport": service.transport,
        "port": service.port, "service_hint": service.service_hint,
        "first_seen_at": timestamp(service.first_seen_at),
        "last_seen_at": timestamp(service.last_seen_at)})
}

fn certificate_json(certificate: Certificate) -> Value {
    json!({"id": certificate.id, "service_id": certificate.service_id,
        "sha256": certificate.sha256, "subject": certificate.subject,
        "issuer": certificate.issuer, "dns_names": certificate.dns_names,
        "not_before": timestamp(certificate.not_before), "not_after": timestamp(certificate.not_after),
        "first_seen_at": timestamp(certificate.first_seen_at),
        "last_seen_at": timestamp(certificate.last_seen_at)})
}

fn website_json(website: Website) -> Value {
    json!({"id": website.id, "service_id": website.service_id, "url": website.url,
        "status_code": website.status_code, "title": website.title, "server": website.server,
        "content_type": website.content_type, "content_sha256": website.content_sha256,
        "first_seen_at": timestamp(website.first_seen_at),
        "last_seen_at": timestamp(website.last_seen_at)})
}

fn schedule_json(schedule: Schedule) -> Value {
    json!({"id": schedule.id, "scope_id": schedule.scope_id, "policy_id": schedule.policy_id,
        "interval_seconds": schedule.interval_seconds, "enabled": schedule.enabled,
        "next_run_at": timestamp(schedule.next_run_at), "last_task_id": schedule.last_task_id,
        "created_at": timestamp(schedule.created_at)})
}

fn asset_change_json(change: AssetChange) -> Value {
    json!({"id": change.id, "schedule_id": change.schedule_id, "task_id": change.task_id,
        "asset_id": change.asset_id, "kind": change.kind,
        "detected_at": timestamp(change.detected_at)})
}

fn exposure_change_json(change: ExposureChange) -> Value {
    json!({"id": change.id, "schedule_id": change.schedule_id, "task_id": change.task_id,
        "resource_kind": change.resource_kind, "resource_id": change.resource_id,
        "kind": change.kind, "previous_fingerprint": change.previous_fingerprint,
        "current_fingerprint": change.current_fingerprint,
        "detected_at": timestamp(change.detected_at)})
}

fn timestamp(value: Option<prost_types::Timestamp>) -> Value {
    value
        .map(|value| json!({"seconds": value.seconds, "nanos": value.nanos}))
        .unwrap_or(Value::Null)
}

struct WebError(crate::RepositoryError);

impl axum::response::IntoResponse for WebError {
    fn into_response(self) -> axum::response::Response {
        let status = match self.0 {
            crate::RepositoryError::NotFound(_) => StatusCode::NOT_FOUND,
            _ => StatusCode::INTERNAL_SERVER_ERROR,
        };
        (
            status,
            Json(json!({"error": status.canonical_reason().unwrap_or("request failed")})),
        )
            .into_response()
    }
}
