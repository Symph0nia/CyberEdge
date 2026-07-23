use std::{
    env,
    net::SocketAddr,
    path::PathBuf,
    sync::Arc,
    time::{Duration, Instant},
};

use axum::{
    Json, Router,
    extract::{Extension, Path, Request, State},
    http::{HeaderName, HeaderValue, StatusCode, header},
    middleware::{self, Next},
    response::{IntoResponse, Response},
    routing::{any, get},
};
use base64::{Engine, engine::general_purpose::STANDARD};
use jsonwebtoken::{Algorithm, DecodingKey, Validation, decode, decode_header, jwk::JwkSet};
use serde_json::{Value, json};
use tokio::sync::RwLock;
use tower_http::{
    services::{ServeDir, ServeFile},
    set_header::SetResponseHeaderLayer,
};

use crate::{
    proto::{
        Asset, AssetChange, Certificate, ExposureChange, Finding, Schedule, Scope, Service, Task,
        Website,
    },
    repository::WebReadRepository,
};

#[derive(Clone)]
struct WebState {
    repository: Arc<dyn WebReadRepository>,
}

#[derive(Clone)]
pub enum WebAccess {
    InsecureLocal,
    Oidc(Arc<OidcAccess>),
}

pub struct OidcAccess {
    issuer: String,
    audience: String,
    role_claim: String,
    read_role: String,
    evidence_role: String,
    sensitive_role: String,
    jwks: RwLock<JwkCache>,
    refresh: Option<JwksRefresh>,
}

struct JwkCache {
    set: JwkSet,
    fetched_at: Instant,
}

struct JwksRefresh {
    client: reqwest::Client,
    url: String,
}

#[derive(Clone)]
struct WebPrincipal {
    roles: Vec<String>,
    sensitive: bool,
}

#[derive(Debug)]
struct WebAuthFailure {
    status: StatusCode,
    message: &'static str,
}

#[derive(Debug, thiserror::Error)]
pub enum WebAccessError {
    #[error(
        "OIDC Web access requires CYBEREDGE_WEB_OIDC_ISSUER, CYBEREDGE_WEB_OIDC_AUDIENCE, and CYBEREDGE_WEB_OIDC_JWKS_URL"
    )]
    IncompleteOidc,
    #[error(
        "unauthenticated Web access requires a loopback bind and CYBEREDGE_WEB_ALLOW_INSECURE_LOCAL=true"
    )]
    UnsafeLocal,
    #[error("OIDC JWKS URL must use HTTPS")]
    InsecureJwks,
    #[error("failed to fetch OIDC JWKS: {0}")]
    JwksFetch(#[from] reqwest::Error),
    #[error("OIDC JWKS exceeds 1 MiB")]
    JwksTooLarge,
    #[error("invalid OIDC JWKS: {0}")]
    InvalidJwks(#[from] serde_json::Error),
}

impl WebAccess {
    pub async fn from_env(address: SocketAddr) -> Result<Self, WebAccessError> {
        let issuer = env::var("CYBEREDGE_WEB_OIDC_ISSUER").ok();
        let audience = env::var("CYBEREDGE_WEB_OIDC_AUDIENCE").ok();
        let jwks_url = env::var("CYBEREDGE_WEB_OIDC_JWKS_URL").ok();
        if issuer.is_some() || audience.is_some() || jwks_url.is_some() {
            let (Some(issuer), Some(audience), Some(jwks_url)) = (issuer, audience, jwks_url)
            else {
                return Err(WebAccessError::IncompleteOidc);
            };
            if !jwks_url.starts_with("https://") {
                return Err(WebAccessError::InsecureJwks);
            }
            let client = web_http_client()?;
            let jwks = fetch_jwks(&client, &jwks_url).await?;
            return Ok(Self::Oidc(Arc::new(
                OidcAccess::new(
                    issuer,
                    audience,
                    env::var("CYBEREDGE_WEB_ROLE_CLAIM").unwrap_or_else(|_| "roles".to_owned()),
                    env::var("CYBEREDGE_WEB_READ_ROLE")
                        .unwrap_or_else(|_| "cyberedge.read".to_owned()),
                    env::var("CYBEREDGE_WEB_EVIDENCE_ROLE")
                        .unwrap_or_else(|_| "cyberedge.evidence.read".to_owned()),
                    env::var("CYBEREDGE_WEB_SENSITIVE_ROLE")
                        .unwrap_or_else(|_| "cyberedge.sensitive.read".to_owned()),
                    jwks,
                )
                .with_refresh(client, jwks_url),
            )));
        }
        if address.ip().is_loopback()
            && env::var("CYBEREDGE_WEB_ALLOW_INSECURE_LOCAL")
                .is_ok_and(|value| value == "true" || value == "1")
        {
            return Ok(Self::InsecureLocal);
        }
        Err(WebAccessError::UnsafeLocal)
    }
}

fn web_http_client() -> Result<reqwest::Client, reqwest::Error> {
    let _ = rustls::crypto::ring::default_provider().install_default();
    reqwest::Client::builder()
        .redirect(reqwest::redirect::Policy::none())
        .timeout(Duration::from_secs(10))
        .build()
}

impl OidcAccess {
    pub fn new(
        issuer: String,
        audience: String,
        role_claim: String,
        read_role: String,
        evidence_role: String,
        sensitive_role: String,
        jwks: JwkSet,
    ) -> Self {
        Self {
            issuer,
            audience,
            role_claim,
            read_role,
            evidence_role,
            sensitive_role,
            jwks: RwLock::new(JwkCache {
                set: jwks,
                fetched_at: Instant::now(),
            }),
            refresh: None,
        }
    }

    fn with_refresh(mut self, client: reqwest::Client, url: String) -> Self {
        self.refresh = Some(JwksRefresh { client, url });
        self
    }

    async fn key(&self, kid: &str) -> Result<Option<jsonwebtoken::jwk::Jwk>, WebAccessError> {
        const MAX_AGE: Duration = Duration::from_secs(15 * 60);
        {
            let cache = self.jwks.read().await;
            if cache.fetched_at.elapsed() < MAX_AGE
                && let Some(key) = cache.set.find(kid)
            {
                return Ok(Some(key.clone()));
            }
        }
        let Some(refresh) = &self.refresh else {
            return Ok(self.jwks.read().await.set.find(kid).cloned());
        };
        let mut cache = self.jwks.write().await;
        if cache.fetched_at.elapsed() < MAX_AGE
            && let Some(key) = cache.set.find(kid)
        {
            return Ok(Some(key.clone()));
        }
        cache.set = fetch_jwks(&refresh.client, &refresh.url).await?;
        cache.fetched_at = Instant::now();
        Ok(cache.set.find(kid).cloned())
    }
}

async fn fetch_jwks(client: &reqwest::Client, url: &str) -> Result<JwkSet, WebAccessError> {
    let mut response = client.get(url).send().await?.error_for_status()?;
    if response
        .content_length()
        .is_some_and(|size| size > 1024 * 1024)
    {
        return Err(WebAccessError::JwksTooLarge);
    }
    let mut bytes = Vec::new();
    while let Some(chunk) = response.chunk().await? {
        if bytes.len() + chunk.len() > 1024 * 1024 {
            return Err(WebAccessError::JwksTooLarge);
        }
        bytes.extend_from_slice(&chunk);
    }
    Ok(serde_json::from_slice(&bytes)?)
}

pub async fn serve_read_only_web(
    repository: Arc<dyn WebReadRepository>,
    address: SocketAddr,
    dist: PathBuf,
    access: WebAccess,
) -> std::io::Result<()> {
    let app = read_only_router(repository, dist, access);
    let listener = tokio::net::TcpListener::bind(address).await?;
    axum::serve(listener, app).await
}

pub fn read_only_router(
    repository: Arc<dyn WebReadRepository>,
    dist: PathBuf,
    access: WebAccess,
) -> Router {
    let index = dist.join("index.html");
    Router::new()
        .route("/api/v1/overview", get(overview))
        .route("/api/v1/scopes/{scope_id}/assets", get(assets))
        .route("/api/v1/scopes/{scope_id}/services", get(services))
        .route("/api/v1/scopes/{scope_id}/certificates", get(certificates))
        .route("/api/v1/scopes/{scope_id}/websites", get(websites))
        .route("/api/v1/scopes/{scope_id}/findings", get(findings))
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
        .layer(middleware::from_fn_with_state(access, authorize_web))
        .with_state(WebState { repository })
}

async fn authorize_web(
    State(access): State<WebAccess>,
    mut request: Request,
    next: Next,
) -> Response {
    let mut principal = match access {
        WebAccess::InsecureLocal => WebPrincipal {
            roles: vec![
                "cyberedge.read".to_owned(),
                "cyberedge.evidence.read".to_owned(),
                "cyberedge.sensitive.read".to_owned(),
            ],
            sensitive: true,
        },
        WebAccess::Oidc(ref config) => match bearer_token(&request) {
            Ok(token) => match oidc_principal(&token, config).await {
                Ok(principal) => principal,
                Err(error) => return web_auth_error(error.status, error.message),
            },
            Err(error) => return web_auth_error(error.status, error.message),
        },
    };
    let (read_role, evidence_role, sensitive_role) = match &access {
        WebAccess::Oidc(config) => (
            config.read_role.as_str(),
            config.evidence_role.as_str(),
            config.sensitive_role.as_str(),
        ),
        WebAccess::InsecureLocal => (
            "cyberedge.read",
            "cyberedge.evidence.read",
            "cyberedge.sensitive.read",
        ),
    };
    if !principal.roles.iter().any(|role| role == read_role) {
        return web_auth_error(StatusCode::FORBIDDEN, "required read-only role is missing");
    }
    if request.uri().path().starts_with("/api/v1/evidence/")
        && !principal.roles.iter().any(|role| role == evidence_role)
    {
        return web_auth_error(StatusCode::FORBIDDEN, "required Evidence role is missing");
    }
    principal.sensitive = principal.roles.iter().any(|role| role == sensitive_role);
    request.extensions_mut().insert(principal);
    next.run(request).await
}

fn bearer_token(request: &Request) -> Result<String, WebAuthFailure> {
    request
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|value| value.to_str().ok())
        .and_then(|value| value.strip_prefix("Bearer "))
        .map(str::to_owned)
        .ok_or(WebAuthFailure {
            status: StatusCode::UNAUTHORIZED,
            message: "Bearer token is required",
        })
}

async fn oidc_principal(value: &str, config: &OidcAccess) -> Result<WebPrincipal, WebAuthFailure> {
    let header = decode_header(value)
        .map_err(|_| auth_failure(StatusCode::UNAUTHORIZED, "invalid token header"))?;
    if header.alg != Algorithm::RS256 {
        return Err(auth_failure(
            StatusCode::UNAUTHORIZED,
            "only RS256 tokens are accepted",
        ));
    }
    let kid = header
        .kid
        .as_deref()
        .ok_or_else(|| auth_failure(StatusCode::UNAUTHORIZED, "token key id is required"))?;
    let jwk = config
        .key(kid)
        .await
        .map_err(|_| auth_failure(StatusCode::SERVICE_UNAVAILABLE, "OIDC key refresh failed"))?
        .ok_or_else(|| auth_failure(StatusCode::UNAUTHORIZED, "token key id is unknown"))?;
    let key = DecodingKey::from_jwk(&jwk)
        .map_err(|_| auth_failure(StatusCode::UNAUTHORIZED, "token key is invalid"))?;
    let mut validation = Validation::new(Algorithm::RS256);
    validation.set_audience(&[&config.audience]);
    validation.set_issuer(&[&config.issuer]);
    validation.set_required_spec_claims(&["exp", "iss", "aud", "sub"]);
    validation.validate_nbf = true;
    let claims = decode::<Value>(&value, &key, &validation)
        .map_err(|_| auth_failure(StatusCode::UNAUTHORIZED, "token validation failed"))?
        .claims;
    let roles = claim_strings(claims.get(&config.role_claim)).ok_or_else(|| {
        auth_failure(
            StatusCode::FORBIDDEN,
            "token role claim is missing or invalid",
        )
    })?;
    Ok(WebPrincipal {
        roles,
        sensitive: false,
    })
}

fn auth_failure(status: StatusCode, message: &'static str) -> WebAuthFailure {
    WebAuthFailure { status, message }
}

fn claim_strings(value: Option<&Value>) -> Option<Vec<String>> {
    match value? {
        Value::String(value) => Some(vec![value.clone()]),
        Value::Array(values) => values
            .iter()
            .map(|value| value.as_str().map(str::to_owned))
            .collect(),
        _ => None,
    }
}

fn web_auth_error(status: StatusCode, message: &'static str) -> Response {
    (status, Json(json!({"error": message}))).into_response()
}

async fn api_not_found() -> (StatusCode, Json<Value>) {
    (
        StatusCode::NOT_FOUND,
        Json(json!({"error": "read-only API route not found"})),
    )
}

async fn overview(
    State(state): State<WebState>,
    Extension(principal): Extension<WebPrincipal>,
) -> Result<Json<Value>, WebError> {
    let model = state.repository.read_overview().await.map_err(WebError)?;
    let show_sensitive = principal.sensitive;
    Ok(Json(json!({
        "counts": {"scopes": model.scope_count, "tasks": model.task_count,
            "assets": model.asset_count, "schedules": model.schedule_count,
            "asset_changes": model.asset_change_count,
            "exposure_changes": model.exposure_change_count,
            "services": model.service_count,
            "certificates": model.certificate_count,
            "websites": model.website_count,
            "findings": model.finding_count,
            "observations": model.observation_count,
            "evidence": model.evidence_count,
            "notifications_pending": model.notification_pending_count,
            "notifications_delivered": model.notification_delivered_count,
            "notifications_dead_letter": model.notification_dead_letter_count},
        "scopes": model.scopes.into_iter().map(|scope| scope_json(scope, show_sensitive)).collect::<Vec<_>>(),
        "tasks": model.tasks.into_iter().map(task_json).collect::<Vec<_>>(),
        "assets": model.assets.into_iter().map(asset_json).collect::<Vec<_>>(),
        "schedules": model.schedules.into_iter().map(schedule_json).collect::<Vec<_>>(),
        "asset_changes": model.asset_changes.into_iter().map(asset_change_json).collect::<Vec<_>>(),
        "exposure_changes": model.exposure_changes.into_iter().map(exposure_change_json).collect::<Vec<_>>(),
        "services": model.services.into_iter().map(service_json).collect::<Vec<_>>(),
        "certificates": model.certificates.into_iter().map(certificate_json).collect::<Vec<_>>(),
        "websites": model.websites.into_iter().map(website_json).collect::<Vec<_>>(),
        "findings": model.findings.into_iter().map(finding_json).collect::<Vec<_>>(),
        "audit_events": model.audit_events.into_iter()
            .map(|event| audit_json(event, show_sensitive)).collect::<Vec<_>>()
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

async fn findings(
    State(state): State<WebState>,
    Path(scope_id): Path<String>,
) -> Result<Json<Value>, WebError> {
    let values = state
        .repository
        .search_findings(&scope_id)
        .await
        .map_err(WebError)?;
    Ok(Json(
        json!({"findings": values.into_iter().map(finding_json).collect::<Vec<_>>() }),
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

fn scope_json(scope: Scope, show_sensitive: bool) -> Value {
    let mut value = json!({"id": scope.id, "name": scope.name,
        "targets": scope.targets.into_iter().map(|target| json!({"kind": target.kind, "value": target.value})).collect::<Vec<_>>(),
        "created_at": timestamp(scope.created_at)});
    if show_sensitive {
        value["authorization_ref"] = Value::String(scope.authorization_ref);
    }
    value
}

fn audit_json(event: crate::proto::AuditEvent, show_sensitive: bool) -> Value {
    let mut value = json!({
        "id": event.id, "operation": event.operation,
        "resource_kind": event.resource_kind, "resource_id": event.resource_id,
        "occurred_at": timestamp(event.occurred_at)
    });
    if show_sensitive {
        value["request_id"] = Value::String(event.request_id);
        value["agent_id"] = Value::String(event.agent_id);
        value["skill_name"] = Value::String(event.skill_name);
        value["skill_version"] = Value::String(event.skill_version);
    }
    value
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
        "fingerprints": website.fingerprints.into_iter().map(|fingerprint| json!({
            "id": fingerprint.id, "name": fingerprint.name, "version": fingerprint.version,
            "detector": fingerprint.detector, "rule_id": fingerprint.rule_id,
            "evidence_id": fingerprint.evidence_id, "cpe_name": fingerprint.cpe_name,
            "cpe_source": fingerprint.cpe_source})).collect::<Vec<_>>(),
        "discovered_paths": website.discovered_paths,
        "screenshot_evidence_id": website.screenshot_evidence_id,
        "first_seen_at": timestamp(website.first_seen_at),
        "last_seen_at": timestamp(website.last_seen_at)})
}

fn finding_json(finding: Finding) -> Value {
    json!({"id": finding.id, "scope_id": finding.scope_id, "task_id": finding.task_id,
        "asset_id": finding.asset_id, "observation_id": finding.observation_id,
        "evidence_id": finding.evidence_id, "detector": finding.detector,
        "rule_id": finding.rule_id, "title": finding.title,
        "description": finding.description, "severity": finding.severity,
        "state": finding.state, "fingerprint": finding.fingerprint,
        "first_seen_at": timestamp(finding.first_seen_at),
        "last_seen_at": timestamp(finding.last_seen_at)})
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

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use axum::{Json, Router, extract::State, routing::get};
    use base64::{Engine, engine::general_purpose::URL_SAFE_NO_PAD};
    use jsonwebtoken::{Algorithm, EncodingKey, Header, encode, jwk::JwkSet};
    use rsa::{RsaPrivateKey, pkcs1::EncodeRsaPrivateKey, traits::PublicKeyParts};
    use serde_json::json;
    use tokio::sync::RwLock;

    use super::{OidcAccess, audit_json, oidc_principal, scope_json, web_http_client};
    use crate::proto::{AuditEvent, Scope, ScopeTarget, TargetKind};

    #[test]
    fn sensitive_scope_and_audit_fields_require_explicit_projection_access() {
        let scope = Scope {
            id: "scope-1".to_owned(),
            name: "Example".to_owned(),
            authorization_ref: "change-secret".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
            created_at: None,
        };
        let audit = AuditEvent {
            id: "audit-1".to_owned(),
            request_id: "request-secret".to_owned(),
            operation: "CreateScope".to_owned(),
            agent_id: "agent-secret".to_owned(),
            skill_name: "skill-secret".to_owned(),
            skill_version: "1.0.0".to_owned(),
            resource_kind: "scope".to_owned(),
            resource_id: "scope-1".to_owned(),
            occurred_at: None,
        };

        let redacted_scope = scope_json(scope.clone(), false);
        let redacted_audit = audit_json(audit.clone(), false);
        assert!(redacted_scope.get("authorization_ref").is_none());
        assert!(redacted_audit.get("request_id").is_none());
        assert!(redacted_audit.get("agent_id").is_none());
        assert!(redacted_audit.get("skill_name").is_none());

        assert_eq!(
            scope_json(scope, true)["authorization_ref"],
            "change-secret"
        );
        assert_eq!(audit_json(audit, true)["agent_id"], "agent-secret");
    }

    #[tokio::test]
    async fn unknown_kid_refreshes_the_bounded_jwks_cache() {
        let first = RsaPrivateKey::new(&mut rand::thread_rng(), 2048).unwrap();
        let second = RsaPrivateKey::new(&mut rand::thread_rng(), 2048).unwrap();
        let initial = jwks("first", &first);
        let current = Arc::new(RwLock::new(jwks("second", &second)));
        let app = Router::new()
            .route("/jwks", get(serve_jwks))
            .with_state(current);
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });
        let config = OidcAccess::new(
            "https://identity.example".to_owned(),
            "cyberedge-web".to_owned(),
            "roles".to_owned(),
            "cyberedge.read".to_owned(),
            "cyberedge.evidence.read".to_owned(),
            "cyberedge.sensitive.read".to_owned(),
            initial,
        )
        .with_refresh(web_http_client().unwrap(), format!("http://{address}/jwks"));
        let mut header = Header::new(Algorithm::RS256);
        header.kid = Some("second".to_owned());
        let token = encode(
            &header,
            &json!({
                "iss": "https://identity.example", "aud": "cyberedge-web",
                "sub": "viewer", "exp": jsonwebtoken::get_current_timestamp() + 300,
                "roles": ["cyberedge.read"]
            }),
            &EncodingKey::from_rsa_der(second.to_pkcs1_der().unwrap().as_bytes()),
        )
        .unwrap();

        let principal = oidc_principal(&token, &config).await.unwrap();
        assert_eq!(principal.roles, vec!["cyberedge.read"]);
        assert!(config.jwks.read().await.set.find("second").is_some());
    }

    async fn serve_jwks(State(jwks): State<Arc<RwLock<JwkSet>>>) -> Json<JwkSet> {
        Json(jwks.read().await.clone())
    }

    fn jwks(kid: &str, key: &RsaPrivateKey) -> JwkSet {
        serde_json::from_value(json!({"keys": [{
            "kty": "RSA", "alg": "RS256", "use": "sig", "kid": kid,
            "n": URL_SAFE_NO_PAD.encode(key.n().to_bytes_be()),
            "e": URL_SAFE_NO_PAD.encode(key.e().to_bytes_be())
        }]}))
        .unwrap()
    }
}
