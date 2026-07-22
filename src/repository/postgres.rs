use std::collections::BTreeSet;

use async_trait::async_trait;
use serde_json::json;
use sqlx::{PgConnection, PgPool, Row, postgres::PgPoolOptions};

use super::{
    ClaimedTask, DiscoveryRecord, Mutation, MutationResult, ReadOverview, Repository,
    RepositoryError,
};
use crate::proto::{
    Asset, AssetChange, AssetChangeKind, AuditEvent, Evidence, Observation, Schedule, Scope,
    ScopeTarget, Service, Task, TaskEvent, TaskState,
};

pub struct PostgresRepository {
    pool: PgPool,
}

impl PostgresRepository {
    pub async fn connect(database_url: &str) -> Result<Self, RepositoryError> {
        let pool = PgPoolOptions::new()
            .max_connections(16)
            .connect(database_url)
            .await?;
        sqlx::migrate!().run(&pool).await?;
        Ok(Self { pool })
    }
}

#[async_trait]
impl Repository for PostgresRepository {
    async fn create_scope(
        &self,
        mutation: &Mutation,
        scope: Scope,
    ) -> Result<MutationResult<Scope>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        lock_idempotency(&mut tx, mutation).await?;
        if let Some(scope_id) = replay(&mut tx, mutation).await? {
            let scope = fetch_scope(&mut tx, &scope_id).await?;
            tx.commit().await?;
            return Ok(MutationResult {
                value: scope,
                event: None,
            });
        }

        let created_at = scope.created_at.expect("scope timestamp must exist");
        sqlx::query(
            "INSERT INTO scopes
             (id, name, authorization_ref, created_at_seconds, created_at_nanos)
             VALUES ($1, $2, $3, $4, $5)",
        )
        .bind(&scope.id)
        .bind(&scope.name)
        .bind(&scope.authorization_ref)
        .bind(created_at.seconds)
        .bind(created_at.nanos)
        .execute(&mut *tx)
        .await?;
        for (position, target) in scope.targets.iter().enumerate() {
            sqlx::query(
                "INSERT INTO scope_targets (scope_id, position, kind, value)
                 VALUES ($1, $2, $3, $4)",
            )
            .bind(&scope.id)
            .bind(position as i32)
            .bind(target.kind)
            .bind(&target.value)
            .execute(&mut *tx)
            .await?;
        }

        record_mutation(&mut tx, mutation, "scope", &scope.id).await?;
        insert_outbox(
            &mut tx,
            "scope",
            &scope.id,
            None,
            "scope.created",
            json!({"scope_id": scope.id}),
        )
        .await?;
        tx.commit().await?;
        Ok(MutationResult {
            value: scope,
            event: None,
        })
    }

    async fn get_scope(&self, scope_id: &str) -> Result<Scope, RepositoryError> {
        let mut connection = self.pool.acquire().await?;
        fetch_scope(&mut connection, scope_id).await
    }

    async fn create_task(
        &self,
        mutation: &Mutation,
        task: Task,
        event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        lock_idempotency(&mut tx, mutation).await?;
        if let Some(task_id) = replay(&mut tx, mutation).await? {
            let task = fetch_task(&mut tx, &task_id).await?;
            tx.commit().await?;
            return Ok(MutationResult {
                value: task,
                event: None,
            });
        }
        ensure_exists(
            &mut tx,
            "SELECT 1 FROM scopes WHERE id = $1",
            &task.scope_id,
            "scope",
        )
        .await?;

        let created_at = task.created_at.expect("task created timestamp must exist");
        let updated_at = task.updated_at.expect("task updated timestamp must exist");
        sqlx::query(
            "INSERT INTO tasks
             (id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
              updated_at_seconds, updated_at_nanos)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
        )
        .bind(&task.id)
        .bind(&task.scope_id)
        .bind(&task.policy_id)
        .bind(task.state)
        .bind(created_at.seconds)
        .bind(created_at.nanos)
        .bind(updated_at.seconds)
        .bind(updated_at.nanos)
        .execute(&mut *tx)
        .await?;
        insert_task_event(&mut tx, &event).await?;
        record_mutation(&mut tx, mutation, "task", &task.id).await?;
        insert_outbox(
            &mut tx,
            "task",
            &task.id,
            Some(event.sequence),
            &event.event_type,
            json!({"task_id": task.id, "state": task.state}),
        )
        .await?;
        tx.commit().await?;
        Ok(MutationResult {
            value: task,
            event: Some(event),
        })
    }

    async fn get_task(&self, task_id: &str) -> Result<Task, RepositoryError> {
        let mut connection = self.pool.acquire().await?;
        fetch_task(&mut connection, task_id).await
    }

    async fn create_schedule(
        &self,
        mutation: &Mutation,
        schedule: Schedule,
    ) -> Result<MutationResult<Schedule>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        lock_idempotency(&mut tx, mutation).await?;
        if let Some(schedule_id) = replay(&mut tx, mutation).await? {
            let schedule = fetch_schedule(&mut tx, &schedule_id).await?;
            tx.commit().await?;
            return Ok(MutationResult {
                value: schedule,
                event: None,
            });
        }
        ensure_exists(
            &mut tx,
            "SELECT 1 FROM scopes WHERE id = $1",
            &schedule.scope_id,
            "scope",
        )
        .await?;
        let next = schedule.next_run_at.expect("schedule next run exists");
        let created = schedule
            .created_at
            .expect("schedule created timestamp exists");
        sqlx::query(
            "INSERT INTO schedules
             (id, scope_id, policy_id, interval_seconds, enabled,
              next_run_at_seconds, next_run_at_nanos, last_task_id,
              created_at_seconds, created_at_nanos)
             VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, $8, $9)",
        )
        .bind(&schedule.id)
        .bind(&schedule.scope_id)
        .bind(&schedule.policy_id)
        .bind(schedule.interval_seconds as i64)
        .bind(schedule.enabled)
        .bind(next.seconds)
        .bind(next.nanos)
        .bind(created.seconds)
        .bind(created.nanos)
        .execute(&mut *tx)
        .await?;
        record_mutation(&mut tx, mutation, "schedule", &schedule.id).await?;
        insert_outbox(
            &mut tx,
            "schedule",
            &schedule.id,
            None,
            "schedule.created",
            json!({"schedule_id": schedule.id}),
        )
        .await?;
        tx.commit().await?;
        Ok(MutationResult {
            value: schedule,
            event: None,
        })
    }

    async fn search_schedules(&self, scope_id: &str) -> Result<Vec<Schedule>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM scopes WHERE id = $1")
            .bind(scope_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("scope"));
        }
        let rows = sqlx::query(
            "SELECT id, scope_id, policy_id, interval_seconds, enabled,
                    next_run_at_seconds, next_run_at_nanos, last_task_id,
                    created_at_seconds, created_at_nanos
             FROM schedules WHERE scope_id = $1 ORDER BY created_at_seconds, created_at_nanos",
        )
        .bind(scope_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(schedule_from_row).collect())
    }

    async fn search_asset_changes(
        &self,
        schedule_id: &str,
    ) -> Result<Vec<AssetChange>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM schedules WHERE id = $1")
            .bind(schedule_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("schedule"));
        }
        let rows = sqlx::query(
            "SELECT id, schedule_id, task_id, asset_id, kind,
                    detected_at_seconds, detected_at_nanos
             FROM asset_changes WHERE schedule_id = $1
             ORDER BY detected_at_seconds DESC, detected_at_nanos DESC, id",
        )
        .bind(schedule_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(asset_change_from_row).collect())
    }

    async fn enqueue_due_schedules(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Vec<Task>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        let rows = sqlx::query(
            "SELECT id, scope_id, policy_id, interval_seconds, enabled,
                    next_run_at_seconds, next_run_at_nanos, last_task_id,
                    created_at_seconds, created_at_nanos
             FROM schedules
             WHERE enabled AND (next_run_at_seconds, next_run_at_nanos) <= ($1, $2)
             ORDER BY next_run_at_seconds, next_run_at_nanos
             FOR UPDATE SKIP LOCKED LIMIT 100",
        )
        .bind(timestamp.seconds)
        .bind(timestamp.nanos)
        .fetch_all(&mut *tx)
        .await?;
        let schedules = rows.iter().map(schedule_from_row).collect::<Vec<_>>();
        let mut tasks = Vec::with_capacity(schedules.len());
        for schedule in schedules {
            let task = Task {
                id: format!("task_{}", uuid::Uuid::now_v7()),
                scope_id: schedule.scope_id.clone(),
                policy_id: schedule.policy_id.clone(),
                state: TaskState::Queued.into(),
                created_at: Some(timestamp),
                updated_at: Some(timestamp),
                schedule_id: schedule.id.clone(),
            };
            sqlx::query(
                "INSERT INTO tasks
                 (id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
                  updated_at_seconds, updated_at_nanos, schedule_id)
                 VALUES ($1, $2, $3, $4, $5, $6, $5, $6, $7)",
            )
            .bind(&task.id)
            .bind(&task.scope_id)
            .bind(&task.policy_id)
            .bind(task.state)
            .bind(timestamp.seconds)
            .bind(timestamp.nanos)
            .bind(&schedule.id)
            .execute(&mut *tx)
            .await?;
            let event = TaskEvent {
                task_id: task.id.clone(),
                sequence: 1,
                event_type: "task.queued".to_owned(),
                occurred_at: Some(timestamp),
            };
            insert_task_event(&mut tx, &event).await?;
            insert_outbox(
                &mut tx,
                "task",
                &task.id,
                Some(1),
                "task.queued",
                json!({"task_id": task.id, "schedule_id": schedule.id}),
            )
            .await?;
            sqlx::query(
                "UPDATE schedules SET last_task_id = $2,
                 next_run_at_seconds = $3, next_run_at_nanos = $4 WHERE id = $1",
            )
            .bind(&schedule.id)
            .bind(&task.id)
            .bind(timestamp.seconds + schedule.interval_seconds as i64)
            .bind(timestamp.nanos)
            .execute(&mut *tx)
            .await?;
            tasks.push(task);
        }
        tx.commit().await?;
        Ok(tasks)
    }

    async fn task_events(
        &self,
        task_id: &str,
        after_sequence: u64,
    ) -> Result<Vec<TaskEvent>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM tasks WHERE id = $1")
            .bind(task_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("task"));
        }
        let rows = sqlx::query(
            "SELECT task_id, sequence, event_type, occurred_at_seconds, occurred_at_nanos
             FROM task_events WHERE task_id = $1 AND sequence > $2 ORDER BY sequence",
        )
        .bind(task_id)
        .bind(after_sequence as i64)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(task_event_from_row).collect())
    }

    async fn cancel_task(
        &self,
        mutation: &Mutation,
        task_id: &str,
        mut event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        lock_idempotency(&mut tx, mutation).await?;
        if let Some(stored_task_id) = replay(&mut tx, mutation).await? {
            let task = fetch_task(&mut tx, &stored_task_id).await?;
            tx.commit().await?;
            return Ok(MutationResult {
                value: task,
                event: None,
            });
        }

        let mut task = fetch_task_for_update(&mut tx, task_id).await?;
        let current = TaskState::try_from(task.state).unwrap_or(TaskState::Unspecified);
        if matches!(
            current,
            TaskState::Completed | TaskState::Failed | TaskState::Canceled
        ) {
            return Err(RepositoryError::TerminalTask);
        }
        let timestamp = event.occurred_at.expect("cancel timestamp must exist");
        let next_sequence: i64 = sqlx::query_scalar(
            "SELECT COALESCE(MAX(sequence), 0) + 1 FROM task_events WHERE task_id = $1",
        )
        .bind(task_id)
        .fetch_one(&mut *tx)
        .await?;
        event.sequence = next_sequence as u64;
        task.state = TaskState::Canceled.into();
        task.updated_at = Some(timestamp);

        sqlx::query(
            "UPDATE tasks SET state = $2, updated_at_seconds = $3, updated_at_nanos = $4
             WHERE id = $1",
        )
        .bind(task_id)
        .bind(task.state)
        .bind(timestamp.seconds)
        .bind(timestamp.nanos)
        .execute(&mut *tx)
        .await?;
        insert_task_event(&mut tx, &event).await?;
        record_mutation(&mut tx, mutation, "task", task_id).await?;
        insert_outbox(
            &mut tx,
            "task",
            task_id,
            Some(event.sequence),
            &event.event_type,
            json!({"task_id": task_id, "state": task.state}),
        )
        .await?;
        tx.commit().await?;
        Ok(MutationResult {
            value: task,
            event: Some(event),
        })
    }

    async fn search_assets(&self, scope_id: &str) -> Result<Vec<Asset>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM scopes WHERE id = $1")
            .bind(scope_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("scope"));
        }
        let rows = sqlx::query(
            "SELECT id, scope_id, kind, value, first_seen_at_seconds, first_seen_at_nanos,
                    last_seen_at_seconds, last_seen_at_nanos
             FROM assets WHERE scope_id = $1 ORDER BY kind, value",
        )
        .bind(scope_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(asset_from_row).collect())
    }

    async fn search_services(&self, scope_id: &str) -> Result<Vec<Service>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM scopes WHERE id = $1")
            .bind(scope_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("scope"));
        }
        let rows = sqlx::query(
            "SELECT services.id, services.asset_id, services.transport, services.port,
                    services.service_hint, services.first_seen_at_seconds,
                    services.first_seen_at_nanos, services.last_seen_at_seconds,
                    services.last_seen_at_nanos
             FROM services JOIN assets ON assets.id = services.asset_id
             WHERE assets.scope_id = $1 ORDER BY assets.value, services.port",
        )
        .bind(scope_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(service_from_row).collect())
    }

    async fn search_observations(
        &self,
        task_id: &str,
    ) -> Result<Vec<Observation>, RepositoryError> {
        let exists = sqlx::query("SELECT 1 FROM tasks WHERE id = $1")
            .bind(task_id)
            .fetch_optional(&self.pool)
            .await?
            .is_some();
        if !exists {
            return Err(RepositoryError::NotFound("task"));
        }
        let rows = sqlx::query(
            "SELECT id, task_id, asset_id, observation_type, value_json::text AS value_json,
                    evidence_id, observed_at_seconds, observed_at_nanos
             FROM observations WHERE task_id = $1 ORDER BY observed_at_seconds, id",
        )
        .bind(task_id)
        .fetch_all(&self.pool)
        .await?;
        Ok(rows.iter().map(observation_from_row).collect())
    }

    async fn get_evidence(&self, evidence_id: &str) -> Result<Evidence, RepositoryError> {
        let row = sqlx::query(
            "SELECT id, media_type, sha256, content, created_at_seconds, created_at_nanos
             FROM evidence WHERE id = $1",
        )
        .bind(evidence_id)
        .fetch_optional(&self.pool)
        .await?
        .ok_or(RepositoryError::NotFound("evidence"))?;
        Ok(evidence_from_row(&row))
    }

    async fn claim_task(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Option<ClaimedTask>, RepositoryError> {
        let mut tx = self.pool.begin().await?;
        let task_id = sqlx::query_scalar::<_, String>(
            "SELECT id FROM tasks WHERE state = $1 ORDER BY created_at_seconds, created_at_nanos
             FOR UPDATE SKIP LOCKED LIMIT 1",
        )
        .bind(i32::from(TaskState::Queued))
        .fetch_optional(&mut *tx)
        .await?;
        let Some(task_id) = task_id else {
            tx.commit().await?;
            return Ok(None);
        };
        sqlx::query(
            "UPDATE tasks SET state = $2, updated_at_seconds = $3, updated_at_nanos = $4
             WHERE id = $1",
        )
        .bind(&task_id)
        .bind(i32::from(TaskState::Running))
        .bind(timestamp.seconds)
        .bind(timestamp.nanos)
        .execute(&mut *tx)
        .await?;
        let event = TaskEvent {
            task_id: task_id.clone(),
            sequence: 2,
            event_type: "task.running".to_owned(),
            occurred_at: Some(timestamp),
        };
        insert_task_event(&mut tx, &event).await?;
        insert_outbox(
            &mut tx,
            "task",
            &task_id,
            Some(2),
            "task.running",
            json!({"task_id": task_id}),
        )
        .await?;
        let task = fetch_task(&mut tx, &task_id).await?;
        let scope = fetch_scope(&mut tx, &task.scope_id).await?;
        tx.commit().await?;
        Ok(Some(ClaimedTask { task, scope }))
    }

    async fn complete_task(
        &self,
        task_id: &str,
        records: Vec<DiscoveryRecord>,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError> {
        let mut tx = self.pool.begin().await?;
        ensure_running(&mut tx, task_id).await?;
        let task = fetch_task(&mut tx, task_id).await?;
        let current_assets = records
            .iter()
            .map(|record| record.asset.id.clone())
            .collect::<BTreeSet<_>>();
        let complete_coverage = records
            .iter()
            .all(|record| !record.observation.observation_type.ends_with(".error"));
        let previous_assets = if task.schedule_id.is_empty() {
            None
        } else {
            let previous_task_id = sqlx::query_scalar::<_, String>(
                "SELECT id FROM tasks
                 WHERE schedule_id = $1 AND id <> $2 AND state = $3
                 ORDER BY updated_at_seconds DESC, updated_at_nanos DESC LIMIT 1",
            )
            .bind(&task.schedule_id)
            .bind(task_id)
            .bind(i32::from(TaskState::Completed))
            .fetch_optional(&mut *tx)
            .await?;
            match previous_task_id {
                Some(previous_task_id) => Some(
                    sqlx::query_scalar::<_, String>(
                        "SELECT DISTINCT asset_id FROM observations WHERE task_id = $1",
                    )
                    .bind(previous_task_id)
                    .fetch_all(&mut *tx)
                    .await?
                    .into_iter()
                    .collect::<BTreeSet<_>>(),
                ),
                None => None,
            }
        };
        for record in records {
            let evidence_at = record
                .evidence
                .created_at
                .expect("evidence timestamp exists");
            sqlx::query(
                "INSERT INTO evidence
                 (id, media_type, sha256, content, created_at_seconds, created_at_nanos)
                 VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (sha256) DO NOTHING",
            )
            .bind(&record.evidence.id)
            .bind(&record.evidence.media_type)
            .bind(&record.evidence.sha256)
            .bind(&record.evidence.content)
            .bind(evidence_at.seconds)
            .bind(evidence_at.nanos)
            .execute(&mut *tx)
            .await?;
            let first_seen = record.asset.first_seen_at.expect("asset timestamp exists");
            let last_seen = record.asset.last_seen_at.expect("asset timestamp exists");
            sqlx::query(
                "INSERT INTO assets
                 (id, scope_id, kind, value, first_seen_at_seconds, first_seen_at_nanos,
                  last_seen_at_seconds, last_seen_at_nanos)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
                 ON CONFLICT (scope_id, kind, value) DO UPDATE SET
                   last_seen_at_seconds = EXCLUDED.last_seen_at_seconds,
                   last_seen_at_nanos = EXCLUDED.last_seen_at_nanos",
            )
            .bind(&record.asset.id)
            .bind(&record.asset.scope_id)
            .bind(record.asset.kind)
            .bind(&record.asset.value)
            .bind(first_seen.seconds)
            .bind(first_seen.nanos)
            .bind(last_seen.seconds)
            .bind(last_seen.nanos)
            .execute(&mut *tx)
            .await?;
            if let Some(service) = record.service {
                let first_seen = service.first_seen_at.expect("service timestamp exists");
                let last_seen = service.last_seen_at.expect("service timestamp exists");
                sqlx::query(
                    "INSERT INTO services
                     (id, asset_id, transport, port, service_hint,
                      first_seen_at_seconds, first_seen_at_nanos,
                      last_seen_at_seconds, last_seen_at_nanos)
                     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                     ON CONFLICT (asset_id, transport, port) DO UPDATE SET
                       service_hint = EXCLUDED.service_hint,
                       last_seen_at_seconds = EXCLUDED.last_seen_at_seconds,
                       last_seen_at_nanos = EXCLUDED.last_seen_at_nanos",
                )
                .bind(&service.id)
                .bind(&service.asset_id)
                .bind(&service.transport)
                .bind(service.port as i32)
                .bind(&service.service_hint)
                .bind(first_seen.seconds)
                .bind(first_seen.nanos)
                .bind(last_seen.seconds)
                .bind(last_seen.nanos)
                .execute(&mut *tx)
                .await?;
            }
            let observed_at = record
                .observation
                .observed_at
                .expect("observation timestamp exists");
            let value_json =
                serde_json::from_str::<serde_json::Value>(&record.observation.value_json)
                    .unwrap_or_else(|_| json!({"raw": record.observation.value_json}));
            sqlx::query(
                "INSERT INTO observations
                 (id, task_id, asset_id, observation_type, value_json, evidence_id,
                  observed_at_seconds, observed_at_nanos)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
            )
            .bind(&record.observation.id)
            .bind(&record.observation.task_id)
            .bind(&record.observation.asset_id)
            .bind(&record.observation.observation_type)
            .bind(value_json)
            .bind(&record.observation.evidence_id)
            .bind(observed_at.seconds)
            .bind(observed_at.nanos)
            .execute(&mut *tx)
            .await?;
        }
        if let Some(previous_assets) = previous_assets {
            for change in asset_changes(
                &task,
                &previous_assets,
                &current_assets,
                timestamp,
                complete_coverage,
            ) {
                sqlx::query(
                    "INSERT INTO asset_changes
                     (id, schedule_id, task_id, asset_id, kind,
                      detected_at_seconds, detected_at_nanos)
                     VALUES ($1, $2, $3, $4, $5, $6, $7)",
                )
                .bind(&change.id)
                .bind(&change.schedule_id)
                .bind(&change.task_id)
                .bind(&change.asset_id)
                .bind(change.kind)
                .bind(timestamp.seconds)
                .bind(timestamp.nanos)
                .execute(&mut *tx)
                .await?;
                insert_outbox(
                    &mut tx,
                    "schedule",
                    &task.schedule_id,
                    None,
                    "monitor.asset_changed",
                    json!({"change_id": change.id, "task_id": task.id,
                        "asset_id": change.asset_id, "kind": change.kind}),
                )
                .await?;
            }
        }
        finish_task(&mut tx, task_id, TaskState::Completed, timestamp).await?;
        tx.commit().await?;
        Ok(())
    }

    async fn fail_task(
        &self,
        task_id: &str,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError> {
        let mut tx = self.pool.begin().await?;
        ensure_running(&mut tx, task_id).await?;
        finish_task(&mut tx, task_id, TaskState::Failed, timestamp).await?;
        tx.commit().await?;
        Ok(())
    }

    async fn read_overview(&self) -> Result<ReadOverview, RepositoryError> {
        let scope_ids = sqlx::query_scalar::<_, String>(
            "SELECT id FROM scopes ORDER BY created_at_seconds DESC, created_at_nanos DESC LIMIT 20",
        )
        .fetch_all(&self.pool)
        .await?;
        let mut scopes = Vec::with_capacity(scope_ids.len());
        for scope_id in scope_ids {
            let mut connection = self.pool.acquire().await?;
            scopes.push(fetch_scope(&mut connection, &scope_id).await?);
        }
        let tasks = sqlx::query(
            "SELECT id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
                    updated_at_seconds, updated_at_nanos, schedule_id FROM tasks
             ORDER BY created_at_seconds DESC, created_at_nanos DESC LIMIT 50",
        )
        .fetch_all(&self.pool)
        .await?
        .iter()
        .map(task_from_row)
        .collect();
        let assets = sqlx::query(
            "SELECT id, scope_id, kind, value, first_seen_at_seconds, first_seen_at_nanos,
                    last_seen_at_seconds, last_seen_at_nanos FROM assets
             ORDER BY last_seen_at_seconds DESC, last_seen_at_nanos DESC LIMIT 100",
        )
        .fetch_all(&self.pool)
        .await?
        .iter()
        .map(asset_from_row)
        .collect();
        let schedules = sqlx::query(
            "SELECT id, scope_id, policy_id, interval_seconds, enabled,
                    next_run_at_seconds, next_run_at_nanos, last_task_id,
                    created_at_seconds, created_at_nanos FROM schedules
             ORDER BY created_at_seconds DESC, created_at_nanos DESC LIMIT 100",
        )
        .fetch_all(&self.pool)
        .await?
        .iter()
        .map(schedule_from_row)
        .collect();
        let asset_changes = sqlx::query(
            "SELECT id, schedule_id, task_id, asset_id, kind,
                    detected_at_seconds, detected_at_nanos FROM asset_changes
             ORDER BY detected_at_seconds DESC, detected_at_nanos DESC LIMIT 100",
        )
        .fetch_all(&self.pool)
        .await?
        .iter()
        .map(asset_change_from_row)
        .collect();
        let services = sqlx::query(
            "SELECT id, asset_id, transport, port, service_hint,
                    first_seen_at_seconds, first_seen_at_nanos,
                    last_seen_at_seconds, last_seen_at_nanos FROM services
             ORDER BY last_seen_at_seconds DESC, last_seen_at_nanos DESC LIMIT 100",
        )
        .fetch_all(&self.pool)
        .await?
        .iter()
        .map(service_from_row)
        .collect();
        let observation_count = sqlx::query_scalar("SELECT COUNT(*) FROM observations")
            .fetch_one(&self.pool)
            .await?;
        let evidence_count = sqlx::query_scalar("SELECT COUNT(*) FROM evidence")
            .fetch_one(&self.pool)
            .await?;
        let scope_count = sqlx::query_scalar("SELECT COUNT(*) FROM scopes")
            .fetch_one(&self.pool)
            .await?;
        let task_count = sqlx::query_scalar("SELECT COUNT(*) FROM tasks")
            .fetch_one(&self.pool)
            .await?;
        let asset_count = sqlx::query_scalar("SELECT COUNT(*) FROM assets")
            .fetch_one(&self.pool)
            .await?;
        let schedule_count = sqlx::query_scalar("SELECT COUNT(*) FROM schedules")
            .fetch_one(&self.pool)
            .await?;
        let asset_change_count = sqlx::query_scalar("SELECT COUNT(*) FROM asset_changes")
            .fetch_one(&self.pool)
            .await?;
        let service_count = sqlx::query_scalar("SELECT COUNT(*) FROM services")
            .fetch_one(&self.pool)
            .await?;
        let audit_events = fetch_audit(&self.pool, 50).await?;
        Ok(ReadOverview {
            scopes,
            tasks,
            assets,
            schedules,
            asset_changes,
            services,
            scope_count,
            task_count,
            asset_count,
            schedule_count,
            asset_change_count,
            service_count,
            observation_count,
            evidence_count,
            audit_events,
        })
    }

    async fn search_audit(&self) -> Result<Vec<AuditEvent>, RepositoryError> {
        fetch_audit(&self.pool, 200).await
    }
}

async fn fetch_audit(pool: &PgPool, limit: i64) -> Result<Vec<AuditEvent>, RepositoryError> {
    let rows = sqlx::query(
        "SELECT id, request_id, operation, agent_id, skill_name, skill_version,
                resource_kind, resource_id, EXTRACT(EPOCH FROM occurred_at)::bigint AS occurred_at_seconds
         FROM audit_events ORDER BY occurred_at DESC LIMIT $1",
    )
    .bind(limit)
    .fetch_all(pool)
    .await?;
    Ok(rows
        .iter()
        .map(|row| AuditEvent {
            id: row.get("id"),
            request_id: row.get("request_id"),
            operation: row.get("operation"),
            agent_id: row.get("agent_id"),
            skill_name: row.get("skill_name"),
            skill_version: row.get("skill_version"),
            resource_kind: row.get("resource_kind"),
            resource_id: row.get("resource_id"),
            occurred_at: Some(prost_types::Timestamp {
                seconds: row.get("occurred_at_seconds"),
                nanos: 0,
            }),
        })
        .collect())
}

async fn ensure_running(
    connection: &mut PgConnection,
    task_id: &str,
) -> Result<(), RepositoryError> {
    let task = fetch_task_for_update(connection, task_id).await?;
    if task.state != i32::from(TaskState::Running) {
        return Err(RepositoryError::TerminalTask);
    }
    Ok(())
}

async fn finish_task(
    connection: &mut PgConnection,
    task_id: &str,
    state: TaskState,
    timestamp: prost_types::Timestamp,
) -> Result<(), RepositoryError> {
    sqlx::query(
        "UPDATE tasks SET state = $2, updated_at_seconds = $3, updated_at_nanos = $4 WHERE id = $1",
    )
    .bind(task_id)
    .bind(i32::from(state))
    .bind(timestamp.seconds)
    .bind(timestamp.nanos)
    .execute(&mut *connection)
    .await?;
    let event_type = match state {
        TaskState::Completed => "task.completed",
        TaskState::Failed => "task.failed",
        _ => unreachable!("only terminal worker states are valid"),
    };
    let event = TaskEvent {
        task_id: task_id.to_owned(),
        sequence: 3,
        event_type: event_type.to_owned(),
        occurred_at: Some(timestamp),
    };
    insert_task_event(connection, &event).await?;
    insert_outbox(
        connection,
        "task",
        task_id,
        Some(3),
        event_type,
        json!({"task_id": task_id, "state": i32::from(state)}),
    )
    .await
}

async fn lock_idempotency(
    connection: &mut PgConnection,
    mutation: &Mutation,
) -> Result<(), RepositoryError> {
    sqlx::query("SELECT pg_advisory_xact_lock(hashtextextended($1, 0))")
        .bind(mutation.key())
        .execute(connection)
        .await?;
    Ok(())
}

async fn replay(
    connection: &mut PgConnection,
    mutation: &Mutation,
) -> Result<Option<String>, RepositoryError> {
    let row = sqlx::query(
        "SELECT request_fingerprint, resource_id FROM idempotency_keys
         WHERE operation = $1 AND agent_id = $2 AND idempotency_key = $3",
    )
    .bind(mutation.operation)
    .bind(&mutation.context.agent_id)
    .bind(&mutation.context.idempotency_key)
    .fetch_optional(connection)
    .await?;
    let Some(row) = row else {
        return Ok(None);
    };
    let fingerprint: Vec<u8> = row.get("request_fingerprint");
    if fingerprint != mutation.fingerprint {
        return Err(RepositoryError::IdempotencyConflict);
    }
    Ok(Some(row.get("resource_id")))
}

async fn record_mutation(
    connection: &mut PgConnection,
    mutation: &Mutation,
    resource_kind: &str,
    resource_id: &str,
) -> Result<(), RepositoryError> {
    sqlx::query(
        "INSERT INTO idempotency_keys
         (operation, agent_id, idempotency_key, request_fingerprint, resource_kind, resource_id)
         VALUES ($1, $2, $3, $4, $5, $6)",
    )
    .bind(mutation.operation)
    .bind(&mutation.context.agent_id)
    .bind(&mutation.context.idempotency_key)
    .bind(&mutation.fingerprint)
    .bind(resource_kind)
    .bind(resource_id)
    .execute(&mut *connection)
    .await?;
    sqlx::query(
        "INSERT INTO audit_events
         (id, request_id, operation, agent_id, skill_name, skill_version, resource_kind, resource_id)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
    )
    .bind(format!("audit_{}", uuid::Uuid::now_v7()))
    .bind(&mutation.context.request_id)
    .bind(mutation.operation)
    .bind(&mutation.context.agent_id)
    .bind(&mutation.context.skill_name)
    .bind(&mutation.context.skill_version)
    .bind(resource_kind)
    .bind(resource_id)
    .execute(connection)
    .await?;
    Ok(())
}

async fn insert_outbox(
    connection: &mut PgConnection,
    aggregate_kind: &str,
    aggregate_id: &str,
    sequence: Option<u64>,
    event_type: &str,
    payload: serde_json::Value,
) -> Result<(), RepositoryError> {
    sqlx::query(
        "INSERT INTO outbox_events
         (id, aggregate_kind, aggregate_id, sequence, event_type, payload)
         VALUES ($1, $2, $3, $4, $5, $6)",
    )
    .bind(format!("event_{}", uuid::Uuid::now_v7()))
    .bind(aggregate_kind)
    .bind(aggregate_id)
    .bind(sequence.map(|value| value as i64))
    .bind(event_type)
    .bind(payload)
    .execute(connection)
    .await?;
    Ok(())
}

async fn insert_task_event(
    connection: &mut PgConnection,
    event: &TaskEvent,
) -> Result<(), RepositoryError> {
    let timestamp = event.occurred_at.expect("event timestamp must exist");
    sqlx::query(
        "INSERT INTO task_events
         (task_id, sequence, event_type, occurred_at_seconds, occurred_at_nanos)
         VALUES ($1, $2, $3, $4, $5)",
    )
    .bind(&event.task_id)
    .bind(event.sequence as i64)
    .bind(&event.event_type)
    .bind(timestamp.seconds)
    .bind(timestamp.nanos)
    .execute(connection)
    .await?;
    Ok(())
}

async fn fetch_scope(
    connection: &mut PgConnection,
    scope_id: &str,
) -> Result<Scope, RepositoryError> {
    let row = sqlx::query(
        "SELECT id, name, authorization_ref, created_at_seconds, created_at_nanos
         FROM scopes WHERE id = $1",
    )
    .bind(scope_id)
    .fetch_optional(&mut *connection)
    .await?
    .ok_or(RepositoryError::NotFound("scope"))?;
    let target_rows =
        sqlx::query("SELECT kind, value FROM scope_targets WHERE scope_id = $1 ORDER BY position")
            .bind(scope_id)
            .fetch_all(connection)
            .await?;
    Ok(Scope {
        id: row.get("id"),
        name: row.get("name"),
        authorization_ref: row.get("authorization_ref"),
        targets: target_rows
            .iter()
            .map(|row| ScopeTarget {
                kind: row.get("kind"),
                value: row.get("value"),
            })
            .collect(),
        created_at: Some(prost_types::Timestamp {
            seconds: row.get("created_at_seconds"),
            nanos: row.get("created_at_nanos"),
        }),
    })
}

async fn fetch_task(connection: &mut PgConnection, task_id: &str) -> Result<Task, RepositoryError> {
    let row = sqlx::query(
        "SELECT id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
                updated_at_seconds, updated_at_nanos, schedule_id FROM tasks WHERE id = $1",
    )
    .bind(task_id)
    .fetch_optional(connection)
    .await?
    .ok_or(RepositoryError::NotFound("task"))?;
    Ok(task_from_row(&row))
}

async fn fetch_schedule(
    connection: &mut PgConnection,
    schedule_id: &str,
) -> Result<Schedule, RepositoryError> {
    let row = sqlx::query(
        "SELECT id, scope_id, policy_id, interval_seconds, enabled,
                next_run_at_seconds, next_run_at_nanos, last_task_id,
                created_at_seconds, created_at_nanos FROM schedules WHERE id = $1",
    )
    .bind(schedule_id)
    .fetch_optional(connection)
    .await?
    .ok_or(RepositoryError::NotFound("schedule"))?;
    Ok(schedule_from_row(&row))
}

async fn fetch_task_for_update(
    connection: &mut PgConnection,
    task_id: &str,
) -> Result<Task, RepositoryError> {
    let row = sqlx::query(
        "SELECT id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
                updated_at_seconds, updated_at_nanos, schedule_id FROM tasks WHERE id = $1 FOR UPDATE",
    )
    .bind(task_id)
    .fetch_optional(connection)
    .await?
    .ok_or(RepositoryError::NotFound("task"))?;
    Ok(task_from_row(&row))
}

async fn ensure_exists(
    connection: &mut PgConnection,
    query: &'static str,
    id: &str,
    kind: &'static str,
) -> Result<(), RepositoryError> {
    if sqlx::query(query)
        .bind(id)
        .fetch_optional(connection)
        .await?
        .is_none()
    {
        return Err(RepositoryError::NotFound(kind));
    }
    Ok(())
}

fn task_from_row(row: &sqlx::postgres::PgRow) -> Task {
    Task {
        id: row.get("id"),
        scope_id: row.get("scope_id"),
        policy_id: row.get("policy_id"),
        state: row.get("state"),
        created_at: Some(prost_types::Timestamp {
            seconds: row.get("created_at_seconds"),
            nanos: row.get("created_at_nanos"),
        }),
        updated_at: Some(prost_types::Timestamp {
            seconds: row.get("updated_at_seconds"),
            nanos: row.get("updated_at_nanos"),
        }),
        schedule_id: row
            .get::<Option<String>, _>("schedule_id")
            .unwrap_or_default(),
    }
}

fn schedule_from_row(row: &sqlx::postgres::PgRow) -> Schedule {
    Schedule {
        id: row.get("id"),
        scope_id: row.get("scope_id"),
        policy_id: row.get("policy_id"),
        interval_seconds: row.get::<i64, _>("interval_seconds") as u64,
        enabled: row.get("enabled"),
        next_run_at: Some(prost_types::Timestamp {
            seconds: row.get("next_run_at_seconds"),
            nanos: row.get("next_run_at_nanos"),
        }),
        last_task_id: row
            .get::<Option<String>, _>("last_task_id")
            .unwrap_or_default(),
        created_at: Some(prost_types::Timestamp {
            seconds: row.get("created_at_seconds"),
            nanos: row.get("created_at_nanos"),
        }),
    }
}

fn asset_change_from_row(row: &sqlx::postgres::PgRow) -> AssetChange {
    AssetChange {
        id: row.get("id"),
        schedule_id: row.get("schedule_id"),
        task_id: row.get("task_id"),
        asset_id: row.get("asset_id"),
        kind: row.get("kind"),
        detected_at: Some(prost_types::Timestamp {
            seconds: row.get("detected_at_seconds"),
            nanos: row.get("detected_at_nanos"),
        }),
    }
}

fn asset_changes(
    task: &Task,
    previous: &BTreeSet<String>,
    current: &BTreeSet<String>,
    timestamp: prost_types::Timestamp,
    include_disappeared: bool,
) -> Vec<AssetChange> {
    current
        .difference(previous)
        .map(|asset_id| (asset_id, AssetChangeKind::Appeared))
        .chain(
            include_disappeared
                .then(|| {
                    previous
                        .difference(current)
                        .map(|asset_id| (asset_id, AssetChangeKind::Disappeared))
                })
                .into_iter()
                .flatten(),
        )
        .map(|(asset_id, kind)| AssetChange {
            id: format!("change_{}", uuid::Uuid::now_v7()),
            schedule_id: task.schedule_id.clone(),
            task_id: task.id.clone(),
            asset_id: asset_id.clone(),
            kind: kind.into(),
            detected_at: Some(timestamp),
        })
        .collect()
}

fn task_event_from_row(row: &sqlx::postgres::PgRow) -> TaskEvent {
    TaskEvent {
        task_id: row.get("task_id"),
        sequence: row.get::<i64, _>("sequence") as u64,
        event_type: row.get("event_type"),
        occurred_at: Some(prost_types::Timestamp {
            seconds: row.get("occurred_at_seconds"),
            nanos: row.get("occurred_at_nanos"),
        }),
    }
}

fn asset_from_row(row: &sqlx::postgres::PgRow) -> Asset {
    Asset {
        id: row.get("id"),
        scope_id: row.get("scope_id"),
        kind: row.get("kind"),
        value: row.get("value"),
        first_seen_at: Some(prost_types::Timestamp {
            seconds: row.get("first_seen_at_seconds"),
            nanos: row.get("first_seen_at_nanos"),
        }),
        last_seen_at: Some(prost_types::Timestamp {
            seconds: row.get("last_seen_at_seconds"),
            nanos: row.get("last_seen_at_nanos"),
        }),
    }
}

fn service_from_row(row: &sqlx::postgres::PgRow) -> Service {
    Service {
        id: row.get("id"),
        asset_id: row.get("asset_id"),
        transport: row.get("transport"),
        port: row.get::<i32, _>("port") as u32,
        service_hint: row.get("service_hint"),
        first_seen_at: Some(prost_types::Timestamp {
            seconds: row.get("first_seen_at_seconds"),
            nanos: row.get("first_seen_at_nanos"),
        }),
        last_seen_at: Some(prost_types::Timestamp {
            seconds: row.get("last_seen_at_seconds"),
            nanos: row.get("last_seen_at_nanos"),
        }),
    }
}

fn observation_from_row(row: &sqlx::postgres::PgRow) -> Observation {
    Observation {
        id: row.get("id"),
        task_id: row.get("task_id"),
        asset_id: row.get("asset_id"),
        observation_type: row.get("observation_type"),
        value_json: row.get("value_json"),
        evidence_id: row.get("evidence_id"),
        observed_at: Some(prost_types::Timestamp {
            seconds: row.get("observed_at_seconds"),
            nanos: row.get("observed_at_nanos"),
        }),
    }
}

fn evidence_from_row(row: &sqlx::postgres::PgRow) -> Evidence {
    Evidence {
        id: row.get("id"),
        media_type: row.get("media_type"),
        sha256: row.get("sha256"),
        content: row.get("content"),
        created_at: Some(prost_types::Timestamp {
            seconds: row.get("created_at_seconds"),
            nanos: row.get("created_at_nanos"),
        }),
    }
}
