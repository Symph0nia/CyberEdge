use std::{
    collections::HashMap,
    net::IpAddr,
    pin::Pin,
    str::FromStr,
    sync::Arc,
    time::{SystemTime, UNIX_EPOCH},
};

use async_stream::try_stream;
use bytes::Bytes;
use prost::Message;
use tokio_stream::Stream;
use tonic::{Code, Request, Response, Status};
use uuid::Uuid;

use crate::proto::{
    AssetChange, CancelTaskRequest, CreateScheduleRequest, CreateScopeRequest, ErrorDetail,
    GetEvidenceRequest, GetScopeRequest, GetTaskReportRequest, GetTaskRequest, HealthResponse,
    InvocationContext, Schedule, Scope, ScopeTarget, SearchAssetChangesRequest,
    SearchAssetChangesResponse, SearchAssetsRequest, SearchAssetsResponse, SearchAuditRequest,
    SearchAuditResponse, SearchCertificatesRequest, SearchCertificatesResponse,
    SearchObservationsRequest, SearchObservationsResponse, SearchSchedulesRequest,
    SearchSchedulesResponse, SearchServicesRequest, SearchServicesResponse, SearchWebsitesRequest,
    SearchWebsitesResponse, StartScanRequest, TargetKind, Task, TaskEvent, TaskReport, TaskState,
    WatchTaskRequest,
    cyber_edge_server::{CyberEdge, CyberEdgeServer},
};
use crate::{
    policy::Authorizer,
    repository::{Mutation, Repository, RepositoryError},
};

#[derive(Clone)]
pub struct CyberEdgeService {
    repository: Arc<dyn Repository>,
    authorizer: Arc<dyn Authorizer>,
}

impl CyberEdgeService {
    pub fn new(repository: Arc<dyn Repository>, authorizer: Arc<dyn Authorizer>) -> Self {
        Self {
            repository,
            authorizer,
        }
    }

    pub fn server(self) -> CyberEdgeServer<Self> {
        CyberEdgeServer::new(self)
    }
}

#[tonic::async_trait]
impl CyberEdge for CyberEdgeService {
    type WatchTaskStream = Pin<Box<dyn Stream<Item = Result<TaskEvent, Status>> + Send + 'static>>;

    async fn health(&self, _request: Request<()>) -> Result<Response<HealthResponse>, Status> {
        Ok(Response::new(HealthResponse {
            status: "ok".to_owned(),
        }))
    }

    async fn create_scope(
        &self,
        request: Request<CreateScopeRequest>,
    ) -> Result<Response<Scope>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "scope.manage")?;
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let mutation = Mutation {
            operation: "scope.create",
            context: context.clone(),
            fingerprint: semantic_request.encode_to_vec(),
        };

        let name = required("name", &request.name)?;
        let authorization_ref = required("authorization_ref", &request.authorization_ref)?;
        if request.targets.is_empty() {
            return Err(invalid("SCOPE_TARGETS_REQUIRED", "scope requires targets"));
        }

        let targets = request
            .targets
            .into_iter()
            .map(normalize_target)
            .collect::<Result<Vec<_>, _>>()?;
        let scope = Scope {
            id: new_id("scope"),
            name: name.to_owned(),
            authorization_ref: authorization_ref.to_owned(),
            targets,
            created_at: Some(now()),
        };

        let result = self
            .repository
            .create_scope(&mutation, scope)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(result.value))
    }

    async fn get_scope(
        &self,
        request: Request<GetScopeRequest>,
    ) -> Result<Response<Scope>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "scope.read")?;
        let scope_id = request.scope_id;
        let scope = self
            .repository
            .get_scope(&scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(scope))
    }

    async fn start_scan(
        &self,
        request: Request<StartScanRequest>,
    ) -> Result<Response<Task>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, policy_capability(&request.policy_id)?)?;
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let mutation = Mutation {
            operation: "scan.start",
            context: context.clone(),
            fingerprint: semantic_request.encode_to_vec(),
        };
        validate_policy(&request.policy_id)?;

        let timestamp = now();
        let task = Task {
            id: new_id("task"),
            scope_id: request.scope_id,
            policy_id: request.policy_id,
            state: TaskState::Queued.into(),
            created_at: Some(timestamp),
            updated_at: Some(timestamp),
            schedule_id: String::new(),
        };
        let event = TaskEvent {
            task_id: task.id.clone(),
            sequence: 1,
            event_type: "task.queued".to_owned(),
            occurred_at: Some(timestamp),
        };
        let result = self
            .repository
            .create_task(&mutation, task, event)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(result.value))
    }

    async fn get_task(&self, request: Request<GetTaskRequest>) -> Result<Response<Task>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "task.read")?;
        let task_id = request.task_id;
        let task = self
            .repository
            .get_task(&task_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(task))
    }

    async fn create_schedule(
        &self,
        request: Request<CreateScheduleRequest>,
    ) -> Result<Response<Schedule>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "schedule.manage")?;
        self.authorize(context, policy_capability(&request.policy_id)?)?;
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let mutation = Mutation {
            operation: "schedule.create",
            context: context.clone(),
            fingerprint: semantic_request.encode_to_vec(),
        };
        validate_policy(&request.policy_id)?;
        if request.interval_seconds < 60 {
            return Err(invalid(
                "SCHEDULE_INTERVAL_INVALID",
                "interval_seconds must be at least 60",
            ));
        }
        let created_at = now();
        let schedule = Schedule {
            id: new_id("schedule"),
            scope_id: required("scope_id", &request.scope_id)?.to_owned(),
            policy_id: request.policy_id,
            interval_seconds: request.interval_seconds,
            enabled: true,
            next_run_at: Some(prost_types::Timestamp {
                seconds: created_at.seconds + request.interval_seconds as i64,
                nanos: created_at.nanos,
            }),
            last_task_id: String::new(),
            created_at: Some(created_at),
        };
        let result = self
            .repository
            .create_schedule(&mutation, schedule)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(result.value))
    }

    async fn search_schedules(
        &self,
        request: Request<SearchSchedulesRequest>,
    ) -> Result<Response<SearchSchedulesResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "schedule.read")?;
        let schedules = self
            .repository
            .search_schedules(&request.scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchSchedulesResponse { schedules }))
    }

    async fn search_asset_changes(
        &self,
        request: Request<SearchAssetChangesRequest>,
    ) -> Result<Response<SearchAssetChangesResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "monitor.read")?;
        let changes: Vec<AssetChange> = self
            .repository
            .search_asset_changes(&request.schedule_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchAssetChangesResponse { changes }))
    }

    async fn watch_task(
        &self,
        request: Request<WatchTaskRequest>,
    ) -> Result<Response<Self::WatchTaskStream>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "task.read")?;
        self.repository
            .get_task(&request.task_id)
            .await
            .map_err(repository_status)?;
        let repository = self.repository.clone();
        let task_id = request.task_id;
        let mut last_sequence = request.after_sequence;

        let stream = try_stream! {
            let mut interval = tokio::time::interval(std::time::Duration::from_millis(250));
            loop {
                interval.tick().await;
                let events = repository
                    .task_events(&task_id, last_sequence)
                    .await
                    .map_err(repository_status)?;
                for event in events {
                    last_sequence = event.sequence;
                    let terminal = matches!(
                        event.event_type.as_str(),
                        "task.completed" | "task.failed" | "task.canceled"
                    );
                    yield event;
                    if terminal {
                        return;
                    }
                }
            }
        };

        Ok(Response::new(Box::pin(stream)))
    }

    async fn cancel_task(
        &self,
        request: Request<CancelTaskRequest>,
    ) -> Result<Response<Task>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "task.cancel")?;
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let mutation = Mutation {
            operation: "task.cancel",
            context: context.clone(),
            fingerprint: semantic_request.encode_to_vec(),
        };
        let timestamp = now();
        let event = TaskEvent {
            task_id: request.task_id.clone(),
            sequence: 0,
            event_type: "task.canceled".to_owned(),
            occurred_at: Some(timestamp),
        };
        let result = self
            .repository
            .cancel_task(&mutation, &request.task_id, event)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(result.value))
    }

    async fn search_assets(
        &self,
        request: Request<SearchAssetsRequest>,
    ) -> Result<Response<SearchAssetsResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "asset.read")?;
        let assets = self
            .repository
            .search_assets(&request.scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchAssetsResponse { assets }))
    }

    async fn search_services(
        &self,
        request: Request<SearchServicesRequest>,
    ) -> Result<Response<SearchServicesResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "service.read")?;
        let services = self
            .repository
            .search_services(&request.scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchServicesResponse { services }))
    }

    async fn search_certificates(
        &self,
        request: Request<SearchCertificatesRequest>,
    ) -> Result<Response<SearchCertificatesResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "certificate.read")?;
        let certificates = self
            .repository
            .search_certificates(&request.scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchCertificatesResponse { certificates }))
    }

    async fn search_websites(
        &self,
        request: Request<SearchWebsitesRequest>,
    ) -> Result<Response<SearchWebsitesResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "website.read")?;
        let websites = self
            .repository
            .search_websites(&request.scope_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchWebsitesResponse { websites }))
    }

    async fn search_observations(
        &self,
        request: Request<SearchObservationsRequest>,
    ) -> Result<Response<SearchObservationsResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "asset.read")?;
        let observations = self
            .repository
            .search_observations(&request.task_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchObservationsResponse { observations }))
    }

    async fn get_evidence(
        &self,
        request: Request<GetEvidenceRequest>,
    ) -> Result<Response<crate::proto::Evidence>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "evidence.read")?;
        let evidence = self
            .repository
            .get_evidence(&request.evidence_id)
            .await
            .map_err(repository_status)?;
        Ok(Response::new(evidence))
    }

    async fn get_task_report(
        &self,
        request: Request<GetTaskReportRequest>,
    ) -> Result<Response<TaskReport>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "report.read")?;
        let task = self
            .repository
            .get_task(&request.task_id)
            .await
            .map_err(repository_status)?;
        if task.state != i32::from(TaskState::Completed) {
            return Err(failed_precondition(
                "REPORT_NOT_READY",
                "task is not completed",
            ));
        }
        let scope = self
            .repository
            .get_scope(&task.scope_id)
            .await
            .map_err(repository_status)?;
        let observations = self
            .repository
            .search_observations(&task.id)
            .await
            .map_err(repository_status)?;
        let asset_ids = observations
            .iter()
            .map(|value| value.asset_id.as_str())
            .collect::<std::collections::HashSet<_>>();
        let assets = self
            .repository
            .search_assets(&scope.id)
            .await
            .map_err(repository_status)?
            .into_iter()
            .filter(|asset| asset_ids.contains(asset.id.as_str()))
            .collect::<Vec<_>>();
        let service_keys = observations
            .iter()
            .filter(|observation| observation.observation_type == "tcp.open")
            .filter_map(|observation| {
                serde_json::from_str::<serde_json::Value>(&observation.value_json)
                    .ok()?
                    .get("port")?
                    .as_u64()
                    .map(|port| (observation.asset_id.as_str(), port as u32))
            })
            .collect::<std::collections::HashSet<_>>();
        let services = self
            .repository
            .search_services(&scope.id)
            .await
            .map_err(repository_status)?
            .into_iter()
            .filter(|service| service_keys.contains(&(service.asset_id.as_str(), service.port)))
            .collect::<Vec<_>>();
        let certificate_hashes = observations
            .iter()
            .filter(|observation| observation.observation_type == "tls.certificate")
            .filter_map(|observation| {
                serde_json::from_str::<serde_json::Value>(&observation.value_json)
                    .ok()?
                    .get("sha256")?
                    .as_str()
                    .map(str::to_owned)
            })
            .collect::<std::collections::HashSet<_>>();
        let certificates = self
            .repository
            .search_certificates(&scope.id)
            .await
            .map_err(repository_status)?
            .into_iter()
            .filter(|certificate| {
                certificate_hashes.contains(&certificate.sha256)
                    && services
                        .iter()
                        .any(|service| service.id == certificate.service_id)
            })
            .collect();
        let website_hashes = observations
            .iter()
            .filter(|observation| observation.observation_type == "http.response")
            .filter_map(|observation| {
                serde_json::from_str::<serde_json::Value>(&observation.value_json)
                    .ok()?
                    .get("content_sha256")?
                    .as_str()
                    .map(str::to_owned)
            })
            .collect::<std::collections::HashSet<_>>();
        let websites = self
            .repository
            .search_websites(&scope.id)
            .await
            .map_err(repository_status)?
            .into_iter()
            .filter(|website| {
                website_hashes.contains(&website.content_sha256)
                    && services
                        .iter()
                        .any(|service| service.id == website.service_id)
            })
            .collect();
        let mut evidence_ids = std::collections::HashSet::new();
        let mut evidence = Vec::new();
        for observation in &observations {
            if evidence_ids.insert(observation.evidence_id.clone()) {
                evidence.push(
                    self.repository
                        .get_evidence(&observation.evidence_id)
                        .await
                        .map_err(repository_status)?,
                );
            }
        }
        Ok(Response::new(TaskReport {
            task: Some(task),
            scope: Some(scope),
            assets,
            observations,
            evidence,
            generated_at: Some(now()),
            services,
            certificates,
            websites,
        }))
    }

    async fn search_audit(
        &self,
        request: Request<SearchAuditRequest>,
    ) -> Result<Response<SearchAuditResponse>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        self.authorize(context, "audit.read")?;
        let events = self
            .repository
            .search_audit()
            .await
            .map_err(repository_status)?;
        Ok(Response::new(SearchAuditResponse { events }))
    }
}

impl CyberEdgeService {
    fn authorize(&self, context: &InvocationContext, capability: &str) -> Result<(), Status> {
        if self.authorizer.authorize(context, capability) {
            return Ok(());
        }
        Err(status(
            Code::PermissionDenied,
            "CAPABILITY_DENIED",
            "agent skill lacks required capability",
            false,
        ))
    }
}

fn validate_context(context: Option<&InvocationContext>) -> Result<&InvocationContext, Status> {
    let context = context.ok_or_else(|| invalid("CONTEXT_REQUIRED", "context is required"))?;
    for (field, value) in [
        ("request_id", context.request_id.as_str()),
        ("idempotency_key", context.idempotency_key.as_str()),
        ("agent_id", context.agent_id.as_str()),
        ("skill_name", context.skill_name.as_str()),
        ("skill_version", context.skill_version.as_str()),
    ] {
        required(field, value)?;
    }
    Ok(context)
}

fn repository_status(error: RepositoryError) -> Status {
    match error {
        RepositoryError::NotFound("scope") => not_found("SCOPE_NOT_FOUND", "scope not found"),
        RepositoryError::NotFound("task") => not_found("TASK_NOT_FOUND", "task not found"),
        RepositoryError::NotFound("evidence") => {
            not_found("EVIDENCE_NOT_FOUND", "evidence not found")
        }
        RepositoryError::NotFound(_) => not_found("RESOURCE_NOT_FOUND", "resource not found"),
        RepositoryError::IdempotencyConflict => failed_precondition(
            "IDEMPOTENCY_KEY_REUSED",
            "idempotency key was already used with different input",
        ),
        RepositoryError::TerminalTask => {
            failed_precondition("TASK_ALREADY_TERMINAL", "terminal task cannot be canceled")
        }
        RepositoryError::Database(error) => {
            eprintln!("repository error: {error}");
            internal("REPOSITORY_FAILURE", "repository operation failed")
        }
        RepositoryError::Migration(error) => {
            eprintln!("migration error: {error}");
            internal("REPOSITORY_FAILURE", "repository migration failed")
        }
    }
}

fn normalize_target(mut target: ScopeTarget) -> Result<ScopeTarget, Status> {
    let kind = TargetKind::try_from(target.kind)
        .map_err(|_| invalid("SCOPE_TARGET_KIND_INVALID", "invalid target kind"))?;
    let value = target.value.trim();
    if value.is_empty() {
        return Err(invalid("SCOPE_TARGET_INVALID", "target value is required"));
    }

    target.value = match kind {
        TargetKind::Domain => normalize_domain(value)?,
        TargetKind::Ip => IpAddr::from_str(value)
            .map_err(|_| invalid("SCOPE_TARGET_INVALID", "invalid IP target"))?
            .to_string(),
        TargetKind::Cidr => normalize_cidr(value)?,
        TargetKind::Organization => value.to_owned(),
        TargetKind::Unspecified => {
            return Err(invalid(
                "SCOPE_TARGET_KIND_REQUIRED",
                "target kind is required",
            ));
        }
    };
    Ok(target)
}

fn normalize_domain(value: &str) -> Result<String, Status> {
    let domain = value.trim_end_matches('.').to_ascii_lowercase();
    let valid = domain.len() <= 253
        && domain.contains('.')
        && domain.split('.').all(|label| {
            !label.is_empty()
                && label.len() <= 63
                && !label.starts_with('-')
                && !label.ends_with('-')
                && label
                    .bytes()
                    .all(|byte| byte.is_ascii_alphanumeric() || byte == b'-')
        });
    if !valid {
        return Err(invalid("SCOPE_TARGET_INVALID", "invalid domain target"));
    }
    Ok(domain)
}

fn normalize_cidr(value: &str) -> Result<String, Status> {
    let (address, prefix) = value
        .split_once('/')
        .ok_or_else(|| invalid("SCOPE_TARGET_INVALID", "invalid CIDR target"))?;
    let address = IpAddr::from_str(address)
        .map_err(|_| invalid("SCOPE_TARGET_INVALID", "invalid CIDR address"))?;
    let prefix = prefix
        .parse::<u8>()
        .map_err(|_| invalid("SCOPE_TARGET_INVALID", "invalid CIDR prefix"))?;
    let maximum = if address.is_ipv4() { 32 } else { 128 };
    if prefix > maximum {
        return Err(invalid("SCOPE_TARGET_INVALID", "invalid CIDR prefix"));
    }
    Ok(format!("{address}/{prefix}"))
}

fn required<'a>(field: &str, value: &'a str) -> Result<&'a str, Status> {
    let value = value.trim();
    if value.is_empty() {
        return Err(invalid("FIELD_REQUIRED", &format!("{field} is required")));
    }
    Ok(value)
}

fn validate_policy(policy_id: &str) -> Result<(), Status> {
    required("policy_id", policy_id)?;
    if matches!(
        policy_id,
        "policy_passive_dns" | "policy_passive_inventory" | "policy_service_baseline"
    ) {
        return Ok(());
    }
    Err(invalid(
        "POLICY_UNSUPPORTED",
        "supported policies: policy_passive_dns, policy_passive_inventory, policy_service_baseline",
    ))
}

fn policy_capability(policy_id: &str) -> Result<&'static str, Status> {
    validate_policy(policy_id)?;
    Ok(match policy_id {
        "policy_service_baseline" => "scan.active",
        _ => "scan.passive",
    })
}

fn new_id(prefix: &str) -> String {
    format!("{prefix}_{}", Uuid::now_v7())
}

fn now() -> prost_types::Timestamp {
    let duration = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("system time must be after Unix epoch");
    prost_types::Timestamp {
        seconds: duration.as_secs() as i64,
        nanos: duration.subsec_nanos() as i32,
    }
}

fn invalid(code: &str, message: &str) -> Status {
    status(Code::InvalidArgument, code, message, false)
}

fn not_found(code: &str, message: &str) -> Status {
    status(Code::NotFound, code, message, false)
}

fn failed_precondition(code: &str, message: &str) -> Status {
    status(Code::FailedPrecondition, code, message, false)
}

fn internal(code: &str, message: &str) -> Status {
    status(Code::Internal, code, message, true)
}

fn status(grpc_code: Code, error_code: &str, message: &str, retryable: bool) -> Status {
    let detail = ErrorDetail {
        code: error_code.to_owned(),
        retryable,
        metadata: HashMap::new(),
    };
    Status::with_details(grpc_code, message, Bytes::from(detail.encode_to_vec()))
}

#[cfg(test)]
#[path = "service_tests.rs"]
mod tests;
