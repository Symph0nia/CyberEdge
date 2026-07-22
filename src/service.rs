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
use tokio::sync::{RwLock, broadcast};
use tokio_stream::Stream;
use tonic::{Code, Request, Response, Status};
use uuid::Uuid;

use crate::proto::{
    CancelTaskRequest, CreateScopeRequest, ErrorDetail, GetScopeRequest, GetTaskRequest,
    HealthResponse, InvocationContext, Scope, ScopeTarget, StartScanRequest, TargetKind, Task,
    TaskEvent, TaskState, WatchTaskRequest,
    cyber_edge_server::{CyberEdge, CyberEdgeServer},
};

#[derive(Default)]
struct State {
    scopes: HashMap<String, Scope>,
    tasks: HashMap<String, Task>,
    events: HashMap<String, Vec<TaskEvent>>,
    event_senders: HashMap<String, broadcast::Sender<TaskEvent>>,
    idempotency: HashMap<String, (Vec<u8>, String)>,
}

#[derive(Clone, Default)]
pub struct CyberEdgeService {
    state: Arc<RwLock<State>>,
}

impl CyberEdgeService {
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
        let idempotency_key = idempotency_key("scope.create", context);
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let fingerprint = semantic_request.encode_to_vec();

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

        let mut state = self.state.write().await;
        if let Some((stored_fingerprint, scope_id)) = state.idempotency.get(&idempotency_key) {
            ensure_same_request(stored_fingerprint, &fingerprint)?;
            let scope = state
                .scopes
                .get(scope_id)
                .expect("idempotent scope must exist")
                .clone();
            return Ok(Response::new(scope));
        }
        state.scopes.insert(scope.id.clone(), scope.clone());
        state
            .idempotency
            .insert(idempotency_key, (fingerprint, scope.id.clone()));
        Ok(Response::new(scope))
    }

    async fn get_scope(
        &self,
        request: Request<GetScopeRequest>,
    ) -> Result<Response<Scope>, Status> {
        let scope_id = request.into_inner().scope_id;
        let state = self.state.read().await;
        let scope = state
            .scopes
            .get(&scope_id)
            .cloned()
            .ok_or_else(|| not_found("SCOPE_NOT_FOUND", "scope not found"))?;
        Ok(Response::new(scope))
    }

    async fn start_scan(
        &self,
        request: Request<StartScanRequest>,
    ) -> Result<Response<Task>, Status> {
        let request = request.into_inner();
        let context = validate_context(request.context.as_ref())?;
        let idempotency_key = idempotency_key("scan.start", context);
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let fingerprint = semantic_request.encode_to_vec();
        required("policy_id", &request.policy_id)?;

        let mut state = self.state.write().await;
        if let Some((stored_fingerprint, task_id)) = state.idempotency.get(&idempotency_key) {
            ensure_same_request(stored_fingerprint, &fingerprint)?;
            let task = state
                .tasks
                .get(task_id)
                .expect("idempotent task must exist")
                .clone();
            return Ok(Response::new(task));
        }
        if !state.scopes.contains_key(&request.scope_id) {
            return Err(not_found("SCOPE_NOT_FOUND", "scope not found"));
        }

        let timestamp = now();
        let task = Task {
            id: new_id("task"),
            scope_id: request.scope_id,
            policy_id: request.policy_id,
            state: TaskState::Queued.into(),
            created_at: Some(timestamp),
            updated_at: Some(timestamp),
        };
        let event = TaskEvent {
            task_id: task.id.clone(),
            sequence: 1,
            event_type: "task.queued".to_owned(),
            occurred_at: Some(timestamp),
        };
        let (sender, _) = broadcast::channel(128);

        state.tasks.insert(task.id.clone(), task.clone());
        state.events.insert(task.id.clone(), vec![event.clone()]);
        state.event_senders.insert(task.id.clone(), sender.clone());
        state
            .idempotency
            .insert(idempotency_key, (fingerprint, task.id.clone()));
        let _ = sender.send(event);

        Ok(Response::new(task))
    }

    async fn get_task(&self, request: Request<GetTaskRequest>) -> Result<Response<Task>, Status> {
        let task_id = request.into_inner().task_id;
        let state = self.state.read().await;
        let task = state
            .tasks
            .get(&task_id)
            .cloned()
            .ok_or_else(|| not_found("TASK_NOT_FOUND", "task not found"))?;
        Ok(Response::new(task))
    }

    async fn watch_task(
        &self,
        request: Request<WatchTaskRequest>,
    ) -> Result<Response<Self::WatchTaskStream>, Status> {
        let request = request.into_inner();
        let state = self.state.read().await;
        let backlog = state
            .events
            .get(&request.task_id)
            .ok_or_else(|| not_found("TASK_NOT_FOUND", "task not found"))?
            .iter()
            .filter(|event| event.sequence > request.after_sequence)
            .cloned()
            .collect::<Vec<_>>();
        let mut receiver = state
            .event_senders
            .get(&request.task_id)
            .expect("task event sender must exist")
            .subscribe();
        drop(state);

        let stream = try_stream! {
            for event in backlog {
                yield event;
            }
            loop {
                match receiver.recv().await {
                    Ok(event) => yield event,
                    Err(broadcast::error::RecvError::Lagged(_)) => {
                        Err(internal("TASK_EVENT_LAGGED", "task event consumer lagged"))?;
                    }
                    Err(broadcast::error::RecvError::Closed) => break,
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
        let idempotency_key = idempotency_key("task.cancel", context);
        let mut semantic_request = request.clone();
        semantic_request.context = None;
        let fingerprint = semantic_request.encode_to_vec();

        let mut state = self.state.write().await;
        if let Some((stored_fingerprint, task_id)) = state.idempotency.get(&idempotency_key) {
            ensure_same_request(stored_fingerprint, &fingerprint)?;
            let task = state
                .tasks
                .get(task_id)
                .expect("idempotent task must exist")
                .clone();
            return Ok(Response::new(task));
        }
        let timestamp = now();
        let task = state
            .tasks
            .get_mut(&request.task_id)
            .ok_or_else(|| not_found("TASK_NOT_FOUND", "task not found"))?;
        let current = TaskState::try_from(task.state).unwrap_or(TaskState::Unspecified);
        if matches!(
            current,
            TaskState::Completed | TaskState::Failed | TaskState::Canceled
        ) {
            return Err(failed_precondition(
                "TASK_ALREADY_TERMINAL",
                "terminal task cannot be canceled",
            ));
        }

        task.state = TaskState::Canceled.into();
        task.updated_at = Some(timestamp);
        let task = task.clone();
        let events = state
            .events
            .get_mut(&request.task_id)
            .expect("task events must exist");
        let event = TaskEvent {
            task_id: request.task_id.clone(),
            sequence: events.len() as u64 + 1,
            event_type: "task.canceled".to_owned(),
            occurred_at: Some(timestamp),
        };
        events.push(event.clone());
        if let Some(sender) = state.event_senders.get(&request.task_id) {
            let _ = sender.send(event);
        }
        state
            .idempotency
            .insert(idempotency_key, (fingerprint, request.task_id));

        Ok(Response::new(task))
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

fn idempotency_key(operation: &str, context: &InvocationContext) -> String {
    format!(
        "{operation}:{}:{}",
        context.agent_id, context.idempotency_key
    )
}

fn ensure_same_request(stored: &[u8], current: &[u8]) -> Result<(), Status> {
    if stored != current {
        return Err(failed_precondition(
            "IDEMPOTENCY_KEY_REUSED",
            "idempotency key was already used with different input",
        ));
    }
    Ok(())
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
