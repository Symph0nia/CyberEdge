use async_trait::async_trait;
use serde_json::json;
use sqlx::{PgConnection, PgPool, Row, postgres::PgPoolOptions};

use super::{Mutation, MutationResult, Repository, RepositoryError};
use crate::proto::{Scope, ScopeTarget, Task, TaskEvent, TaskState};

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
                updated_at_seconds, updated_at_nanos FROM tasks WHERE id = $1",
    )
    .bind(task_id)
    .fetch_optional(connection)
    .await?
    .ok_or(RepositoryError::NotFound("task"))?;
    Ok(task_from_row(&row))
}

async fn fetch_task_for_update(
    connection: &mut PgConnection,
    task_id: &str,
) -> Result<Task, RepositoryError> {
    let row = sqlx::query(
        "SELECT id, scope_id, policy_id, state, created_at_seconds, created_at_nanos,
                updated_at_seconds, updated_at_nanos FROM tasks WHERE id = $1 FOR UPDATE",
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
    }
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
