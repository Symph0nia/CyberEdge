use std::{
    collections::HashMap,
    time::{SystemTime, UNIX_EPOCH},
};

use async_trait::async_trait;
use tokio::sync::RwLock;

use super::{
    ClaimedTask, DiscoveryRecord, Mutation, MutationResult, ReadOverview, Repository,
    RepositoryError,
};
use crate::proto::{
    Asset, AuditEvent, Evidence, Observation, Schedule, Scope, Task, TaskEvent, TaskState,
};

#[derive(Default)]
struct State {
    scopes: HashMap<String, Scope>,
    tasks: HashMap<String, Task>,
    schedules: HashMap<String, Schedule>,
    events: HashMap<String, Vec<TaskEvent>>,
    idempotency: HashMap<String, (Vec<u8>, String)>,
    assets: HashMap<String, Asset>,
    observations: HashMap<String, Observation>,
    evidence: HashMap<String, Evidence>,
    audits: Vec<AuditEvent>,
}

#[derive(Default)]
pub struct MemoryRepository {
    state: RwLock<State>,
}

#[async_trait]
impl Repository for MemoryRepository {
    async fn create_scope(
        &self,
        mutation: &Mutation,
        scope: Scope,
    ) -> Result<MutationResult<Scope>, RepositoryError> {
        let mut state = self.state.write().await;
        if let Some((fingerprint, scope_id)) = state.idempotency.get(&mutation.key()) {
            ensure_same(fingerprint, &mutation.fingerprint)?;
            return Ok(MutationResult {
                value: state.scopes[scope_id].clone(),
                event: None,
            });
        }

        state.scopes.insert(scope.id.clone(), scope.clone());
        state.idempotency.insert(
            mutation.key(),
            (mutation.fingerprint.clone(), scope.id.clone()),
        );
        state.audits.push(audit_event(mutation, "scope", &scope.id));
        Ok(MutationResult {
            value: scope,
            event: None,
        })
    }

    async fn get_scope(&self, scope_id: &str) -> Result<Scope, RepositoryError> {
        self.state
            .read()
            .await
            .scopes
            .get(scope_id)
            .cloned()
            .ok_or(RepositoryError::NotFound("scope"))
    }

    async fn create_task(
        &self,
        mutation: &Mutation,
        task: Task,
        event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError> {
        let mut state = self.state.write().await;
        if let Some((fingerprint, task_id)) = state.idempotency.get(&mutation.key()) {
            ensure_same(fingerprint, &mutation.fingerprint)?;
            return Ok(MutationResult {
                value: state.tasks[task_id].clone(),
                event: None,
            });
        }
        if !state.scopes.contains_key(&task.scope_id) {
            return Err(RepositoryError::NotFound("scope"));
        }

        state.tasks.insert(task.id.clone(), task.clone());
        state.events.insert(task.id.clone(), vec![event.clone()]);
        state.idempotency.insert(
            mutation.key(),
            (mutation.fingerprint.clone(), task.id.clone()),
        );
        state.audits.push(audit_event(mutation, "task", &task.id));
        Ok(MutationResult {
            value: task,
            event: Some(event),
        })
    }

    async fn get_task(&self, task_id: &str) -> Result<Task, RepositoryError> {
        self.state
            .read()
            .await
            .tasks
            .get(task_id)
            .cloned()
            .ok_or(RepositoryError::NotFound("task"))
    }

    async fn create_schedule(
        &self,
        mutation: &Mutation,
        schedule: Schedule,
    ) -> Result<MutationResult<Schedule>, RepositoryError> {
        let mut state = self.state.write().await;
        if let Some((fingerprint, schedule_id)) = state.idempotency.get(&mutation.key()) {
            ensure_same(fingerprint, &mutation.fingerprint)?;
            return Ok(MutationResult {
                value: state.schedules[schedule_id].clone(),
                event: None,
            });
        }
        if !state.scopes.contains_key(&schedule.scope_id) {
            return Err(RepositoryError::NotFound("scope"));
        }
        state
            .schedules
            .insert(schedule.id.clone(), schedule.clone());
        state.idempotency.insert(
            mutation.key(),
            (mutation.fingerprint.clone(), schedule.id.clone()),
        );
        state
            .audits
            .push(audit_event(mutation, "schedule", &schedule.id));
        Ok(MutationResult {
            value: schedule,
            event: None,
        })
    }

    async fn search_schedules(&self, scope_id: &str) -> Result<Vec<Schedule>, RepositoryError> {
        let state = self.state.read().await;
        if !state.scopes.contains_key(scope_id) {
            return Err(RepositoryError::NotFound("scope"));
        }
        Ok(state
            .schedules
            .values()
            .filter(|schedule| schedule.scope_id == scope_id)
            .cloned()
            .collect())
    }

    async fn enqueue_due_schedules(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Vec<Task>, RepositoryError> {
        let mut state = self.state.write().await;
        let due = state
            .schedules
            .values()
            .filter(|schedule| {
                schedule.enabled
                    && schedule.next_run_at.is_some_and(|next| {
                        (next.seconds, next.nanos) <= (timestamp.seconds, timestamp.nanos)
                    })
            })
            .map(|schedule| schedule.id.clone())
            .collect::<Vec<_>>();
        let mut tasks = Vec::with_capacity(due.len());
        for schedule_id in due {
            let schedule = state.schedules.get_mut(&schedule_id).unwrap();
            let task = scheduled_task(schedule, timestamp);
            schedule.last_task_id = task.id.clone();
            schedule.next_run_at = Some(prost_types::Timestamp {
                seconds: timestamp.seconds + schedule.interval_seconds as i64,
                nanos: timestamp.nanos,
            });
            state.events.insert(
                task.id.clone(),
                vec![TaskEvent {
                    task_id: task.id.clone(),
                    sequence: 1,
                    event_type: "task.queued".to_owned(),
                    occurred_at: Some(timestamp),
                }],
            );
            state.tasks.insert(task.id.clone(), task.clone());
            tasks.push(task);
        }
        Ok(tasks)
    }

    async fn task_events(
        &self,
        task_id: &str,
        after_sequence: u64,
    ) -> Result<Vec<TaskEvent>, RepositoryError> {
        self.state
            .read()
            .await
            .events
            .get(task_id)
            .map(|events| {
                events
                    .iter()
                    .filter(|event| event.sequence > after_sequence)
                    .cloned()
                    .collect()
            })
            .ok_or(RepositoryError::NotFound("task"))
    }

    async fn cancel_task(
        &self,
        mutation: &Mutation,
        task_id: &str,
        mut event: TaskEvent,
    ) -> Result<MutationResult<Task>, RepositoryError> {
        let mut state = self.state.write().await;
        if let Some((fingerprint, stored_task_id)) = state.idempotency.get(&mutation.key()) {
            ensure_same(fingerprint, &mutation.fingerprint)?;
            return Ok(MutationResult {
                value: state.tasks[stored_task_id].clone(),
                event: None,
            });
        }

        let task = state
            .tasks
            .get_mut(task_id)
            .ok_or(RepositoryError::NotFound("task"))?;
        let current = TaskState::try_from(task.state).unwrap_or(TaskState::Unspecified);
        if matches!(
            current,
            TaskState::Completed | TaskState::Failed | TaskState::Canceled
        ) {
            return Err(RepositoryError::TerminalTask);
        }
        task.state = TaskState::Canceled.into();
        task.updated_at = event.occurred_at;
        let task = task.clone();

        let events = state
            .events
            .get_mut(task_id)
            .expect("task events must exist");
        event.sequence = events.len() as u64 + 1;
        events.push(event.clone());
        state.idempotency.insert(
            mutation.key(),
            (mutation.fingerprint.clone(), task_id.to_owned()),
        );
        state.audits.push(audit_event(mutation, "task", task_id));
        Ok(MutationResult {
            value: task,
            event: Some(event),
        })
    }

    async fn search_assets(&self, scope_id: &str) -> Result<Vec<Asset>, RepositoryError> {
        let state = self.state.read().await;
        if !state.scopes.contains_key(scope_id) {
            return Err(RepositoryError::NotFound("scope"));
        }
        Ok(state
            .assets
            .values()
            .filter(|asset| asset.scope_id == scope_id)
            .cloned()
            .collect())
    }

    async fn search_observations(
        &self,
        task_id: &str,
    ) -> Result<Vec<Observation>, RepositoryError> {
        let state = self.state.read().await;
        if !state.tasks.contains_key(task_id) {
            return Err(RepositoryError::NotFound("task"));
        }
        Ok(state
            .observations
            .values()
            .filter(|observation| observation.task_id == task_id)
            .cloned()
            .collect())
    }

    async fn get_evidence(&self, evidence_id: &str) -> Result<Evidence, RepositoryError> {
        self.state
            .read()
            .await
            .evidence
            .get(evidence_id)
            .cloned()
            .ok_or(RepositoryError::NotFound("evidence"))
    }

    async fn claim_task(
        &self,
        timestamp: prost_types::Timestamp,
    ) -> Result<Option<ClaimedTask>, RepositoryError> {
        let mut state = self.state.write().await;
        let Some(task_id) = state
            .tasks
            .values()
            .filter(|task| task.state == i32::from(TaskState::Queued))
            .min_by_key(|task| {
                task.created_at
                    .map(|value| (value.seconds, value.nanos))
                    .unwrap_or_default()
            })
            .map(|task| task.id.clone())
        else {
            return Ok(None);
        };
        let task = state.tasks.get_mut(&task_id).expect("selected task exists");
        task.state = TaskState::Running.into();
        task.updated_at = Some(timestamp);
        let task = task.clone();
        state.events.get_mut(&task_id).unwrap().push(TaskEvent {
            task_id,
            sequence: 2,
            event_type: "task.running".to_owned(),
            occurred_at: Some(timestamp),
        });
        let scope = state.scopes[&task.scope_id].clone();
        Ok(Some(ClaimedTask { task, scope }))
    }

    async fn complete_task(
        &self,
        task_id: &str,
        records: Vec<DiscoveryRecord>,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError> {
        let mut state = self.state.write().await;
        finish_memory_task(&mut state, task_id, TaskState::Completed, timestamp)?;
        for record in records {
            state.assets.insert(record.asset.id.clone(), record.asset);
            state
                .observations
                .insert(record.observation.id.clone(), record.observation);
            state
                .evidence
                .insert(record.evidence.id.clone(), record.evidence);
        }
        Ok(())
    }

    async fn fail_task(
        &self,
        task_id: &str,
        timestamp: prost_types::Timestamp,
    ) -> Result<(), RepositoryError> {
        let mut state = self.state.write().await;
        finish_memory_task(&mut state, task_id, TaskState::Failed, timestamp)
    }

    async fn read_overview(&self) -> Result<ReadOverview, RepositoryError> {
        let state = self.state.read().await;
        Ok(ReadOverview {
            scopes: state.scopes.values().cloned().collect(),
            tasks: state.tasks.values().cloned().collect(),
            assets: state.assets.values().cloned().collect(),
            schedules: state.schedules.values().cloned().collect(),
            scope_count: state.scopes.len() as i64,
            task_count: state.tasks.len() as i64,
            asset_count: state.assets.len() as i64,
            schedule_count: state.schedules.len() as i64,
            observation_count: state.observations.len() as i64,
            evidence_count: state.evidence.len() as i64,
            audit_events: state.audits.iter().rev().take(50).cloned().collect(),
        })
    }

    async fn search_audit(&self) -> Result<Vec<AuditEvent>, RepositoryError> {
        Ok(self
            .state
            .read()
            .await
            .audits
            .iter()
            .rev()
            .take(200)
            .cloned()
            .collect())
    }
}

fn audit_event(mutation: &Mutation, resource_kind: &str, resource_id: &str) -> AuditEvent {
    let duration = SystemTime::now().duration_since(UNIX_EPOCH).unwrap();
    AuditEvent {
        id: format!("audit_{}", uuid::Uuid::now_v7()),
        request_id: mutation.context.request_id.clone(),
        operation: mutation.operation.to_owned(),
        agent_id: mutation.context.agent_id.clone(),
        skill_name: mutation.context.skill_name.clone(),
        skill_version: mutation.context.skill_version.clone(),
        resource_kind: resource_kind.to_owned(),
        resource_id: resource_id.to_owned(),
        occurred_at: Some(prost_types::Timestamp {
            seconds: duration.as_secs() as i64,
            nanos: duration.subsec_nanos() as i32,
        }),
    }
}

fn scheduled_task(schedule: &Schedule, timestamp: prost_types::Timestamp) -> Task {
    Task {
        id: format!("task_{}", uuid::Uuid::now_v7()),
        scope_id: schedule.scope_id.clone(),
        policy_id: schedule.policy_id.clone(),
        state: TaskState::Queued.into(),
        created_at: Some(timestamp),
        updated_at: Some(timestamp),
    }
}

fn finish_memory_task(
    state: &mut State,
    task_id: &str,
    target: TaskState,
    timestamp: prost_types::Timestamp,
) -> Result<(), RepositoryError> {
    let task = state
        .tasks
        .get_mut(task_id)
        .ok_or(RepositoryError::NotFound("task"))?;
    if task.state != i32::from(TaskState::Running) {
        return Err(RepositoryError::TerminalTask);
    }
    task.state = target.into();
    task.updated_at = Some(timestamp);
    let events = state.events.get_mut(task_id).unwrap();
    events.push(TaskEvent {
        task_id: task_id.to_owned(),
        sequence: events.len() as u64 + 1,
        event_type: format!(
            "task.{}",
            target
                .as_str_name()
                .to_ascii_lowercase()
                .replace("task_state_", "")
        ),
        occurred_at: Some(timestamp),
    });
    Ok(())
}

fn ensure_same(stored: &[u8], current: &[u8]) -> Result<(), RepositoryError> {
    if stored != current {
        return Err(RepositoryError::IdempotencyConflict);
    }
    Ok(())
}
