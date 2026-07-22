mod memory;
mod postgres;

use async_trait::async_trait;

use crate::proto::{
    Asset, AssetChange, AuditEvent, Certificate, Evidence, InvocationContext, Observation,
    Schedule, Scope, Service, Task, TaskEvent, Website,
};

pub use memory::MemoryRepository;
pub use postgres::PostgresRepository;

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
}

pub struct ReadOverview {
    pub scopes: Vec<Scope>,
    pub tasks: Vec<Task>,
    pub assets: Vec<Asset>,
    pub schedules: Vec<Schedule>,
    pub asset_changes: Vec<AssetChange>,
    pub services: Vec<Service>,
    pub certificates: Vec<Certificate>,
    pub websites: Vec<Website>,
    pub scope_count: i64,
    pub task_count: i64,
    pub asset_count: i64,
    pub schedule_count: i64,
    pub asset_change_count: i64,
    pub service_count: i64,
    pub certificate_count: i64,
    pub website_count: i64,
    pub observation_count: i64,
    pub evidence_count: i64,
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
