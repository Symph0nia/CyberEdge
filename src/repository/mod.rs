mod memory;
mod postgres;

use async_trait::async_trait;
use sha2::{Digest, Sha256};
use std::collections::{BTreeMap, BTreeSet};

use crate::proto::{
    Asset, AssetChange, AuditEvent, Certificate, Evidence, ExposureChange, Finding,
    InvocationContext, Observation, Schedule, Scope, Service, Task, TaskEvent, Website,
};

pub use memory::MemoryRepository;
pub use postgres::{OutboxEvent, PostgresRepository};

#[derive(Clone)]
pub struct Mutation {
    pub operation: &'static str,
    pub context: InvocationContext,
    pub fingerprint: Vec<u8>,
}

impl Mutation {
    pub fn key(&self) -> String {
        format!(
            "{}:{}:{}",
            self.operation, self.context.agent_id, self.context.idempotency_key
        )
    }
}

pub struct MutationResult<T> {
    pub value: T,
    pub event: Option<TaskEvent>,
}

pub struct ClaimedTask {
    pub task: Task,
    pub scope: Scope,
}

pub struct DiscoveryRecord {
    pub asset: Asset,
    pub service: Option<Service>,
    pub certificate: Option<Certificate>,
    pub website: Option<Website>,
    pub observation: Observation,
    pub evidence: Evidence,
    pub findings: Vec<Finding>,
}

pub struct ReadOverview {
    pub scopes: Vec<Scope>,
    pub tasks: Vec<Task>,
    pub assets: Vec<Asset>,
    pub schedules: Vec<Schedule>,
    pub asset_changes: Vec<AssetChange>,
    pub exposure_changes: Vec<ExposureChange>,
    pub services: Vec<Service>,
    pub certificates: Vec<Certificate>,
    pub websites: Vec<Website>,
    pub findings: Vec<Finding>,
    pub scope_count: i64,
    pub task_count: i64,
    pub asset_count: i64,
    pub schedule_count: i64,
    pub asset_change_count: i64,
    pub exposure_change_count: i64,
    pub service_count: i64,
    pub certificate_count: i64,
    pub website_count: i64,
    pub finding_count: i64,
    pub observation_count: i64,
    pub evidence_count: i64,
    pub notification_pending_count: i64,
    pub notification_delivered_count: i64,
    pub notification_dead_letter_count: i64,
    pub audit_events: Vec<AuditEvent>,
}

#[derive(Debug, thiserror::Error)]
pub enum RepositoryError {
    #[error("{0} not found")]
    NotFound(&'static str),
    #[error("idempotency key reused with different input")]
    IdempotencyConflict,
    #[error("task is already terminal")]
    TerminalTask,
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),
    #[error("migration error: {0}")]
    Migration(#[from] sqlx::migrate::MigrateError),
}

#[async_trait]
pub trait Repository: Send + Sync {
    async fn create_scope(
        &self,
        mutation: &Mutation,
        scope: Scope,
    ) -> Result<MutationResult<Scope>, RepositoryError>;

    async fn get_scope(&self, scope_id: &str) -> Result<Scope, RepositoryError>;

    async fn create_task(
        &self,
        mutation: &Mutation,
        task: Task,
        event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError>;

    async fn get_task(&self, task_id: &str) -> Result<Task, RepositoryError>;

    async fn create_schedule(
        &self,
        mutation: &Mutation,
        schedule: Schedule,
    ) -> Result<MutationResult<Schedule>, RepositoryError>;

    async fn search_schedules(&self, scope_id: &str) -> Result<Vec<Schedule>, RepositoryError>;

    async fn search_asset_changes(
        &self,
        schedule_id: &str,
    ) -> Result<Vec<AssetChange>, RepositoryError>;

    async fn search_exposure_changes(
        &self,
        schedule_id: &str,
    ) -> Result<Vec<ExposureChange>, RepositoryError>;

    async fn enqueue_due_schedules(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Vec<Task>, RepositoryError>;

    async fn task_events(
        &self,
        task_id: &str,
        after_sequence: u64,
    ) -> Result<Vec<TaskEvent>, RepositoryError>;

    async fn cancel_task(
        &self,
        mutation: &Mutation,
        task_id: &str,
        event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError>;

    async fn search_assets(&self, scope_id: &str) -> Result<Vec<Asset>, RepositoryError>;

    async fn search_services(&self, scope_id: &str) -> Result<Vec<Service>, RepositoryError>;

    async fn search_certificates(
        &self,
        scope_id: &str,
    ) -> Result<Vec<Certificate>, RepositoryError>;

    async fn search_websites(&self, scope_id: &str) -> Result<Vec<Website>, RepositoryError>;

    async fn report_finding(
        &self,
        mutation: &Mutation,
        finding: Finding,
    ) -> Result<MutationResult<Finding>, RepositoryError>;

    async fn search_findings(&self, scope_id: &str) -> Result<Vec<Finding>, RepositoryError>;

    async fn search_observations(&self, task_id: &str)
    -> Result<Vec<Observation>, RepositoryError>;

    async fn get_evidence(&self, evidence_id: &str) -> Result<Evidence, RepositoryError>;

    async fn claim_task(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Option<ClaimedTask>, RepositoryError>;

    async fn complete_task(
        &self,
        task_id: &str,
        records: Vec<DiscoveryRecord>,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError>;

    async fn fail_task(
        &self,
        task_id: &str,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError>;

    async fn read_overview(&self) -> Result<ReadOverview, RepositoryError>;

    async fn search_audit(&self) -> Result<Vec<AuditEvent>, RepositoryError>;
}

#[derive(Clone)]
struct ExposureState {
    resource_kind: &'static str,
    fingerprint: String,
}

fn exposure_snapshot<'a>(
    observations: impl IntoIterator<Item = &'a Observation>,
) -> BTreeMap<String, ExposureState> {
    observations
        .into_iter()
        .filter_map(|observation| {
            let value = serde_json::from_str::<serde_json::Value>(&observation.value_json).ok()?;
            let fingerprint = digest(&serde_json::to_vec(&value).ok()?);
            let (resource_kind, resource_id) = match observation.observation_type.as_str() {
                "tcp.open" => {
                    let port = value.get("port")?.as_u64()?;
                    let service_id = content_id(
                        "service",
                        format!("{}:tcp:{port}", observation.asset_id).as_bytes(),
                    );
                    ("service", service_id)
                }
                "http.response" => {
                    let port = value.get("port")?.as_u64()?;
                    let service_id = content_id(
                        "service",
                        format!("{}:tcp:{port}", observation.asset_id).as_bytes(),
                    );
                    ("website", content_id("website", service_id.as_bytes()))
                }
                _ => return None,
            };
            Some((
                resource_id,
                ExposureState {
                    resource_kind,
                    fingerprint,
                },
            ))
        })
        .collect()
}

fn exposure_changes(
    task: &Task,
    previous: &BTreeMap<String, ExposureState>,
    current: &BTreeMap<String, ExposureState>,
    timestamp: prost_types::Timestamp,
    include_disappeared: bool,
) -> Vec<ExposureChange> {
    let keys = previous
        .keys()
        .chain(current.keys())
        .cloned()
        .collect::<BTreeSet<_>>();
    keys.into_iter()
        .filter_map(|resource_id| {
            let before = previous.get(&resource_id);
            let after = current.get(&resource_id);
            let kind = match (before, after) {
                (None, Some(_)) => crate::proto::ExposureChangeKind::Appeared,
                (Some(_), None) if include_disappeared => {
                    crate::proto::ExposureChangeKind::Disappeared
                }
                (Some(before), Some(after)) if before.fingerprint != after.fingerprint => {
                    crate::proto::ExposureChangeKind::Modified
                }
                _ => return None,
            };
            let state = after.or(before).expect("change has one state");
            Some(ExposureChange {
                id: format!("exposure_change_{}", uuid::Uuid::now_v7()),
                schedule_id: task.schedule_id.clone(),
                task_id: task.id.clone(),
                resource_kind: state.resource_kind.to_owned(),
                resource_id,
                kind: kind.into(),
                previous_fingerprint: before
                    .map(|state| state.fingerprint.clone())
                    .unwrap_or_default(),
                current_fingerprint: after
                    .map(|state| state.fingerprint.clone())
                    .unwrap_or_default(),
                detected_at: Some(timestamp),
            })
        })
        .collect()
}

fn content_id(prefix: &str, value: &[u8]) -> String {
    format!("{prefix}_{}", &digest(value)[..32])
}

fn digest(value: &[u8]) -> String {
    Sha256::digest(value)
        .iter()
        .map(|byte| format!("{byte:02x}"))
        .collect()
}
