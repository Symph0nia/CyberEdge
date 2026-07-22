use std::sync::Arc;

use cyberedge::{
    Authorizer, CyberEdgeService, PostgresRepository,
    proto::{
        CreateScopeRequest, GetTaskRequest, InvocationContext, ScopeTarget, StartScanRequest,
        TargetKind, WatchTaskRequest, cyber_edge_server::CyberEdge,
    },
};
use sqlx::PgPool;
use tokio_stream::StreamExt;
use tonic::Request;

struct TestAuthorizer;

impl Authorizer for TestAuthorizer {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

fn authorizer() -> Arc<dyn Authorizer> {
    Arc::new(TestAuthorizer)
}

#[tokio::test]
async fn persists_scope_task_events_audit_and_outbox() {
    let Ok(database_url) = std::env::var("TEST_DATABASE_URL") else {
        return;
    };
    let repository = PostgresRepository::connect(&database_url).await.unwrap();
    let pool = PgPool::connect(&database_url).await.unwrap();
    reset(&pool).await;

    let service = CyberEdgeService::new(Arc::new(repository), authorizer());
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("scope")),
            name: "Persistent scope".to_owned(),
            authorization_ref: "authorization:postgres-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "Example.COM".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("task")),
            scope_id: scope.id,
            policy_id: "policy_passive".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    drop(service);

    let restarted = CyberEdgeService::new(
        Arc::new(PostgresRepository::connect(&database_url).await.unwrap()),
        authorizer(),
    );
    let stored = restarted
        .get_task(Request::new(GetTaskRequest {
            context: Some(context("get")),
            task_id: task.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner();
    let mut events = restarted
        .watch_task(Request::new(WatchTaskRequest {
            context: Some(context("watch")),
            task_id: task.id,
            after_sequence: 0,
        }))
        .await
        .unwrap()
        .into_inner();

    assert_eq!(stored.policy_id, "policy_passive");
    assert_eq!(
        events.next().await.unwrap().unwrap().event_type,
        "task.queued"
    );
    let audit_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM audit_events")
        .fetch_one(&pool)
        .await
        .unwrap();
    let outbox_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM outbox_events")
        .fetch_one(&pool)
        .await
        .unwrap();
    assert_eq!(audit_count, 2);
    assert_eq!(outbox_count, 2);
}

fn context(suffix: &str) -> InvocationContext {
    InvocationContext {
        request_id: format!("req_{suffix}"),
        idempotency_key: format!("idem_{suffix}"),
        agent_id: "agent_postgres_test".to_owned(),
        skill_name: "asset-discovery".to_owned(),
        skill_version: "1.0.0".to_owned(),
    }
}

async fn reset(pool: &PgPool) {
    sqlx::query(
        "TRUNCATE outbox_events, audit_events, idempotency_keys, task_events, tasks,
         scope_targets, scopes RESTART IDENTITY CASCADE",
    )
    .execute(pool)
    .await
    .unwrap();
}
