use std::{
    env,
    io::{self, Read},
    path::PathBuf,
};

use base64::{Engine, engine::general_purpose::STANDARD};
use cyberedge::proto::{
    CancelTaskRequest, CreateScheduleRequest, CreateScopeRequest, ErrorDetail, GetEvidenceRequest,
    GetScopeRequest, GetTaskReportRequest, GetTaskRequest, InvocationContext, ReportFindingRequest,
    ScopeTarget, SearchAssetChangesRequest, SearchAssetsRequest, SearchAuditRequest,
    SearchCertificatesRequest, SearchExposureChangesRequest, SearchFindingsRequest,
    SearchObservationsRequest, SearchSchedulesRequest, SearchServicesRequest,
    SearchWebsitesRequest, StartScanRequest, TargetKind, WatchTaskRequest,
    cyber_edge_client::CyberEdgeClient,
};
use hyper_util::rt::TokioIo;
use prost::Message;
use serde::Deserialize;
use serde_json::{Value, json};
use tokio::net::UnixStream;
use tokio_stream::StreamExt;
use tonic::{
    Code, Status,
    transport::{Certificate, Channel, ClientTlsConfig, Endpoint, Identity, Uri},
};
use tower::service_fn;

#[derive(Deserialize)]
struct Envelope {
    request_id: String,
    idempotency_key: String,
    agent_id: String,
    skill_name: String,
    skill_version: String,
    #[serde(flatten)]
    command: Command,
}

#[derive(Deserialize)]
#[serde(tag = "action", rename_all = "snake_case")]
enum Command {
    CreateScope {
        name: String,
        authorization_ref: String,
        targets: Vec<TargetInput>,
    },
    GetScope {
        scope_id: String,
    },
    StartScan {
        scope_id: String,
        policy_id: String,
    },
    GetTask {
        task_id: String,
    },
    WatchTask {
        task_id: String,
        #[serde(default)]
        after_sequence: u64,
    },
    CancelTask {
        task_id: String,
    },
    CreateSchedule {
        scope_id: String,
        policy_id: String,
        interval_seconds: u64,
    },
    SearchSchedules {
        scope_id: String,
    },
    SearchAssetChanges {
        schedule_id: String,
    },
    SearchExposureChanges {
        schedule_id: String,
    },
    SearchAssets {
        scope_id: String,
    },
    SearchServices {
        scope_id: String,
    },
    SearchCertificates {
        scope_id: String,
    },
    SearchWebsites {
        scope_id: String,
    },
    ReportFinding {
        task_id: String,
        observation_id: String,
        detector: String,
        rule_id: String,
        title: String,
        description: String,
        severity: i32,
        fingerprint: String,
    },
    SearchFindings {
        scope_id: String,
    },
    SearchObservations {
        task_id: String,
    },
    GetEvidence {
        evidence_id: String,
    },
    GetTaskReport {
        task_id: String,
    },
    SearchAudit,
}

#[derive(Deserialize)]
struct TargetInput {
    kind: String,
    value: String,
}

#[tokio::main]
async fn main() {
    if let Err(error) = run().await {
        println!("{}", json!({"ok": false, "error": error}));
        std::process::exit(1);
    }
}

async fn run() -> Result<(), Value> {
    let mut input = String::new();
    io::stdin()
        .read_to_string(&mut input)
        .map_err(local_error)?;
    let envelope: Envelope = serde_json::from_str(&input).map_err(local_error)?;
    let channel = connect().await?;
    let mut client = CyberEdgeClient::new(channel);
    let context = Some(InvocationContext {
        request_id: envelope.request_id,
        idempotency_key: envelope.idempotency_key,
        agent_id: envelope.agent_id,
        skill_name: envelope.skill_name,
        skill_version: envelope.skill_version,
    });

    match envelope.command {
        Command::CreateScope {
            name,
            authorization_ref,
            targets,
        } => {
            let targets = targets.into_iter().map(target).collect::<Result<_, _>>()?;
            let value = client
                .create_scope(CreateScopeRequest {
                    context,
                    name,
                    authorization_ref,
                    targets,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(scope_json(value));
        }
        Command::GetScope { scope_id } => {
            let value = client
                .get_scope(GetScopeRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(scope_json(value));
        }
        Command::StartScan {
            scope_id,
            policy_id,
        } => {
            let value = client
                .start_scan(StartScanRequest {
                    context,
                    scope_id,
                    policy_id,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(task_json(value));
        }
        Command::GetTask { task_id } => {
            let value = client
                .get_task(GetTaskRequest { context, task_id })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(task_json(value));
        }
        Command::WatchTask {
            task_id,
            after_sequence,
        } => {
            let mut stream = client
                .watch_task(WatchTaskRequest {
                    context,
                    task_id,
                    after_sequence,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            while let Some(event) = stream.next().await {
                let event = event.map_err(rpc_error)?;
                emit(json!({"task_id": event.task_id, "sequence": event.sequence,
                    "event_type": event.event_type, "occurred_at": timestamp_json(event.occurred_at)}));
            }
        }
        Command::CancelTask { task_id } => {
            let value = client
                .cancel_task(CancelTaskRequest { context, task_id })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(task_json(value));
        }
        Command::CreateSchedule {
            scope_id,
            policy_id,
            interval_seconds,
        } => {
            let value = client
                .create_schedule(CreateScheduleRequest {
                    context,
                    scope_id,
                    policy_id,
                    interval_seconds,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(schedule_json(value));
        }
        Command::SearchSchedules { scope_id } => {
            let values = client
                .search_schedules(SearchSchedulesRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .schedules
                .into_iter()
                .map(schedule_json)
                .collect::<Vec<_>>();
            emit(json!({"schedules": values}));
        }
        Command::SearchAssetChanges { schedule_id } => {
            let values = client
                .search_asset_changes(SearchAssetChangesRequest {
                    context,
                    schedule_id,
                })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .changes
                .into_iter()
                .map(|change| {
                    json!({
                        "id": change.id, "schedule_id": change.schedule_id,
                        "task_id": change.task_id, "asset_id": change.asset_id,
                        "kind": change.kind, "detected_at": timestamp_json(change.detected_at)
                    })
                })
                .collect::<Vec<_>>();
            emit(json!({"changes": values}));
        }
        Command::SearchExposureChanges { schedule_id } => {
            let values = client
                .search_exposure_changes(SearchExposureChangesRequest {
                    context,
                    schedule_id,
                })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .changes
                .into_iter()
                .map(|change| {
                    json!({
                        "id": change.id, "schedule_id": change.schedule_id,
                        "task_id": change.task_id, "resource_kind": change.resource_kind,
                        "resource_id": change.resource_id, "kind": change.kind,
                        "previous_fingerprint": change.previous_fingerprint,
                        "current_fingerprint": change.current_fingerprint,
                        "detected_at": timestamp_json(change.detected_at)
                    })
                })
                .collect::<Vec<_>>();
            emit(json!({"changes": values}));
        }
        Command::SearchAssets { scope_id } => {
            let values = client
                .search_assets(SearchAssetsRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .assets
                .into_iter()
                .map(|asset| {
                    json!({
                        "id": asset.id, "scope_id": asset.scope_id, "kind": asset.kind,
                        "value": asset.value, "first_seen_at": timestamp_json(asset.first_seen_at),
                        "last_seen_at": timestamp_json(asset.last_seen_at)
                    })
                })
                .collect::<Vec<_>>();
            emit(json!({"assets": values}));
        }
        Command::SearchServices { scope_id } => {
            let values = client
                .search_services(SearchServicesRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .services
                .into_iter()
                .map(|service| {
                    json!({
                        "id": service.id, "asset_id": service.asset_id,
                        "transport": service.transport, "port": service.port,
                        "service_hint": service.service_hint,
                        "first_seen_at": timestamp_json(service.first_seen_at),
                        "last_seen_at": timestamp_json(service.last_seen_at)
                    })
                })
                .collect::<Vec<_>>();
            emit(json!({"services": values}));
        }
        Command::SearchCertificates { scope_id } => {
            let values = client
                .search_certificates(SearchCertificatesRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .certificates
                .into_iter()
                .map(certificate_json)
                .collect::<Vec<_>>();
            emit(json!({"certificates": values}));
        }
        Command::SearchWebsites { scope_id } => {
            let values = client
                .search_websites(SearchWebsitesRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .websites
                .into_iter()
                .map(website_json)
                .collect::<Vec<_>>();
            emit(json!({"websites": values}));
        }
        Command::ReportFinding {
            task_id,
            observation_id,
            detector,
            rule_id,
            title,
            description,
            severity,
            fingerprint,
        } => {
            let value = client
                .report_finding(ReportFindingRequest {
                    context,
                    task_id,
                    observation_id,
                    detector,
                    rule_id,
                    title,
                    description,
                    severity,
                    fingerprint,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(finding_json(value));
        }
        Command::SearchFindings { scope_id } => {
            let values = client
                .search_findings(SearchFindingsRequest { context, scope_id })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .findings
                .into_iter()
                .map(finding_json)
                .collect::<Vec<_>>();
            emit(json!({"findings": values}));
        }
        Command::SearchObservations { task_id } => {
            let values = client.search_observations(SearchObservationsRequest { context, task_id })
                .await.map_err(rpc_error)?.into_inner().observations.into_iter().map(|item| json!({
                    "id": item.id, "task_id": item.task_id, "asset_id": item.asset_id,
                    "observation_type": item.observation_type,
                    "value": serde_json::from_str::<Value>(&item.value_json).unwrap_or(Value::String(item.value_json)),
                    "evidence_id": item.evidence_id, "observed_at": timestamp_json(item.observed_at)
                })).collect::<Vec<_>>();
            emit(json!({"observations": values}));
        }
        Command::GetEvidence { evidence_id } => {
            let item = client
                .get_evidence(GetEvidenceRequest {
                    context,
                    evidence_id,
                })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(
                json!({"id": item.id, "media_type": item.media_type, "sha256": item.sha256,
                "content_base64": STANDARD.encode(item.content), "created_at": timestamp_json(item.created_at)}),
            );
        }
        Command::GetTaskReport { task_id } => {
            let report = client
                .get_task_report(GetTaskReportRequest { context, task_id })
                .await
                .map_err(rpc_error)?
                .into_inner();
            emit(json!({
                "task": report.task.map(task_json),
                "scope": report.scope.map(scope_json),
                "assets": report.assets.into_iter().map(asset_json).collect::<Vec<_>>(),
                "observations": report.observations.into_iter().map(observation_json).collect::<Vec<_>>(),
                "evidence": report.evidence.into_iter().map(evidence_json).collect::<Vec<_>>(),
                "services": report.services.into_iter().map(service_json).collect::<Vec<_>>(),
                "certificates": report.certificates.into_iter().map(certificate_json).collect::<Vec<_>>(),
                "websites": report.websites.into_iter().map(website_json).collect::<Vec<_>>(),
                "findings": report.findings.into_iter().map(finding_json).collect::<Vec<_>>(),
                "generated_at": timestamp_json(report.generated_at)
            }));
        }
        Command::SearchAudit => {
            let events = client
                .search_audit(SearchAuditRequest { context })
                .await
                .map_err(rpc_error)?
                .into_inner()
                .events
                .into_iter()
                .map(|event| json!({
                    "id": event.id, "request_id": event.request_id, "operation": event.operation,
                    "agent_id": event.agent_id, "skill_name": event.skill_name,
                    "skill_version": event.skill_version, "resource_kind": event.resource_kind,
                    "resource_id": event.resource_id, "occurred_at": timestamp_json(event.occurred_at)
                }))
                .collect::<Vec<_>>();
            emit(json!({"events": events}));
        }
    }
    Ok(())
}

async fn connect() -> Result<Channel, Value> {
    if let Ok(endpoint) = env::var("CYBEREDGE_RPC_ENDPOINT") {
        let domain = env::var("CYBEREDGE_TLS_DOMAIN").map_err(local_error)?;
        let ca = Certificate::from_pem(read_env_file("CYBEREDGE_TLS_CA")?);
        let identity = Identity::from_pem(
            read_env_file("CYBEREDGE_TLS_CERT")?,
            read_env_file("CYBEREDGE_TLS_KEY")?,
        );
        return Endpoint::from_shared(endpoint)
            .map_err(local_error)?
            .tls_config(
                ClientTlsConfig::new()
                    .domain_name(domain)
                    .ca_certificate(ca)
                    .identity(identity),
            )
            .map_err(local_error)?
            .connect()
            .await
            .map_err(local_error);
    }
    let socket = PathBuf::from(
        env::var("CYBEREDGE_RPC_SOCKET").unwrap_or_else(|_| "/tmp/cyberedge.sock".to_owned()),
    );
    Endpoint::try_from("http://[::]:50051")
        .map_err(local_error)?
        .connect_with_connector(service_fn(move |_: Uri| {
            let socket = socket.clone();
            async move { UnixStream::connect(socket).await.map(TokioIo::new) }
        }))
        .await
        .map_err(local_error)
}

fn read_env_file(variable: &str) -> Result<Vec<u8>, Value> {
    let path = env::var(variable).map_err(local_error)?;
    std::fs::read(path).map_err(local_error)
}

fn target(input: TargetInput) -> Result<ScopeTarget, Value> {
    let kind = match input.kind.as_str() {
        "domain" => TargetKind::Domain,
        "ip" => TargetKind::Ip,
        "cidr" => TargetKind::Cidr,
        "organization" => TargetKind::Organization,
        _ => return Err(json!({"code": "TARGET_KIND_INVALID", "retryable": false})),
    };
    Ok(ScopeTarget {
        kind: kind.into(),
        value: input.value,
    })
}

fn scope_json(value: cyberedge::proto::Scope) -> Value {
    json!({"id": value.id, "name": value.name, "authorization_ref": value.authorization_ref,
        "targets": value.targets.into_iter().map(|target| json!({"kind": target.kind, "value": target.value})).collect::<Vec<_>>(),
        "created_at": timestamp_json(value.created_at)})
}

fn task_json(value: cyberedge::proto::Task) -> Value {
    json!({"id": value.id, "scope_id": value.scope_id, "policy_id": value.policy_id,
        "state": value.state, "created_at": timestamp_json(value.created_at),
        "updated_at": timestamp_json(value.updated_at), "schedule_id": value.schedule_id})
}

fn schedule_json(value: cyberedge::proto::Schedule) -> Value {
    json!({"id": value.id, "scope_id": value.scope_id, "policy_id": value.policy_id,
        "interval_seconds": value.interval_seconds, "enabled": value.enabled,
        "next_run_at": timestamp_json(value.next_run_at), "last_task_id": value.last_task_id,
        "created_at": timestamp_json(value.created_at)})
}

fn asset_json(value: cyberedge::proto::Asset) -> Value {
    json!({"id": value.id, "scope_id": value.scope_id, "kind": value.kind, "value": value.value,
        "first_seen_at": timestamp_json(value.first_seen_at), "last_seen_at": timestamp_json(value.last_seen_at)})
}

fn observation_json(value: cyberedge::proto::Observation) -> Value {
    json!({"id": value.id, "task_id": value.task_id, "asset_id": value.asset_id,
        "observation_type": value.observation_type,
        "value": serde_json::from_str::<Value>(&value.value_json).unwrap_or(Value::String(value.value_json)),
        "evidence_id": value.evidence_id, "observed_at": timestamp_json(value.observed_at)})
}

fn evidence_json(value: cyberedge::proto::Evidence) -> Value {
    json!({"id": value.id, "media_type": value.media_type, "sha256": value.sha256,
        "content_base64": STANDARD.encode(value.content), "created_at": timestamp_json(value.created_at)})
}

fn service_json(value: cyberedge::proto::Service) -> Value {
    json!({"id": value.id, "asset_id": value.asset_id, "transport": value.transport,
        "port": value.port, "service_hint": value.service_hint,
        "first_seen_at": timestamp_json(value.first_seen_at),
        "last_seen_at": timestamp_json(value.last_seen_at)})
}

fn certificate_json(value: cyberedge::proto::Certificate) -> Value {
    json!({"id": value.id, "service_id": value.service_id, "sha256": value.sha256,
        "subject": value.subject, "issuer": value.issuer, "dns_names": value.dns_names,
        "not_before": timestamp_json(value.not_before), "not_after": timestamp_json(value.not_after),
        "first_seen_at": timestamp_json(value.first_seen_at),
        "last_seen_at": timestamp_json(value.last_seen_at)})
}

fn website_json(value: cyberedge::proto::Website) -> Value {
    json!({"id": value.id, "service_id": value.service_id, "url": value.url,
        "status_code": value.status_code, "title": value.title, "server": value.server,
        "content_type": value.content_type, "content_sha256": value.content_sha256,
        "first_seen_at": timestamp_json(value.first_seen_at),
        "last_seen_at": timestamp_json(value.last_seen_at)})
}

fn finding_json(value: cyberedge::proto::Finding) -> Value {
    json!({"id": value.id, "scope_id": value.scope_id, "task_id": value.task_id,
        "asset_id": value.asset_id, "observation_id": value.observation_id,
        "evidence_id": value.evidence_id, "detector": value.detector,
        "rule_id": value.rule_id, "title": value.title, "description": value.description,
        "severity": value.severity, "state": value.state, "fingerprint": value.fingerprint,
        "first_seen_at": timestamp_json(value.first_seen_at),
        "last_seen_at": timestamp_json(value.last_seen_at)})
}

fn timestamp_json(value: Option<prost_types::Timestamp>) -> Value {
    value
        .map(|value| json!({"seconds": value.seconds, "nanos": value.nanos}))
        .unwrap_or(Value::Null)
}

fn emit(value: Value) {
    println!("{}", json!({"ok": true, "result": value}));
}

fn rpc_error(status: Status) -> Value {
    let detail = ErrorDetail::decode(status.details()).ok();
    json!({"grpc_code": code_name(status.code()), "message": status.message(),
        "code": detail.as_ref().map(|value| value.code.as_str()).unwrap_or("RPC_ERROR"),
        "retryable": detail.as_ref().is_some_and(|value| value.retryable),
        "metadata": detail.map(|value| value.metadata).unwrap_or_default()})
}

fn code_name(code: Code) -> &'static str {
    match code {
        Code::Ok => "OK",
        Code::Cancelled => "CANCELLED",
        Code::Unknown => "UNKNOWN",
        Code::InvalidArgument => "INVALID_ARGUMENT",
        Code::DeadlineExceeded => "DEADLINE_EXCEEDED",
        Code::NotFound => "NOT_FOUND",
        Code::AlreadyExists => "ALREADY_EXISTS",
        Code::PermissionDenied => "PERMISSION_DENIED",
        Code::ResourceExhausted => "RESOURCE_EXHAUSTED",
        Code::FailedPrecondition => "FAILED_PRECONDITION",
        Code::Aborted => "ABORTED",
        Code::OutOfRange => "OUT_OF_RANGE",
        Code::Unimplemented => "UNIMPLEMENTED",
        Code::Internal => "INTERNAL",
        Code::Unavailable => "UNAVAILABLE",
        Code::DataLoss => "DATA_LOSS",
        Code::Unauthenticated => "UNAUTHENTICATED",
    }
}

fn local_error(error: impl std::fmt::Display) -> Value {
    json!({"code": "BRIDGE_ERROR", "retryable": false, "message": error.to_string()})
}
