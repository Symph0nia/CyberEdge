use std::collections::HashMap;

use async_trait::async_trait;
use tokio::sync::RwLock;

use super::{Mutation, MutationResult, Repository, RepositoryError};
use crate::proto::{Scope, Task, TaskEvent, TaskState};

#[derive(Default)]
struct State {
    scopes: HashMap<String, Scope>,
    tasks: HashMap<String, Task>,
    events: HashMap<String, Vec<TaskEvent>>,
    idempotency: HashMap<String, (Vec<u8>, String)>,
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
        Ok(MutationResult {
            value: task,
            event: Some(event),
        })
    }
}

fn ensure_same(stored: &[u8], current: &[u8]) -> Result<(), RepositoryError> {
    if stored != current {
        return Err(RepositoryError::IdempotencyConflict);
    }
    Ok(())
}
