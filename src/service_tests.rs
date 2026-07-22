use std::time::Duration;

use tokio_stream::StreamExt;

use super::*;
use crate::MemoryRepository;

struct TestAuthorizer;

impl Authorizer for TestAuthorizer {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

fn service() -> CyberEdgeService {
    CyberEdgeService::new(
        Arc::new(MemoryRepository::default()),
        Arc::new(TestAuthorizer),
    )
}

fn context() -> InvocationContext {
    InvocationContext {
        request_id: "req_test".to_owned(),
        idempotency_key: "idem_test".to_owned(),
        agent_id: "agent_test".to_owned(),
        skill_name: "asset-discovery".to_owned(),
        skill_version: "1.0.0".to_owned(),
    }
}

async fn create_scope(service: &CyberEdgeService) -> Scope {
    service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context()),
            name: "Example".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "Example.COM.".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner()
}

#[tokio::test]
async fn creates_normalized_scope() {
    let service = service();
    let scope = create_scope(&service).await;

    assert!(scope.id.starts_with("scope_"));
    assert_eq!(scope.targets[0].value, "example.com");
}

#[tokio::test]
async fn scope_creation_is_idempotent() {
    let service = service();
    let first = create_scope(&service).await;
    let second = create_scope(&service).await;

    assert_eq!(first.id, second.id);
}

#[tokio::test]
async fn rejects_idempotency_key_reuse_with_different_input() {
    let service = service();
    create_scope(&service).await;
    let error = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context()),
            name: "Different".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "different.example".to_owned(),
            }],
        }))
        .await
        .unwrap_err();

    assert_eq!(error.code(), Code::FailedPrecondition);
    let detail = ErrorDetail::decode(error.details()).unwrap();
    assert_eq!(detail.code, "IDEMPOTENCY_KEY_REUSED");
}

#[tokio::test]
async fn rejects_scope_without_authorization() {
    let service = service();
    let error = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context()),
            name: "Example".to_owned(),
            authorization_ref: String::new(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap_err();

    assert_eq!(error.code(), Code::InvalidArgument);
}

#[tokio::test]
async fn streams_and_cancels_task() {
    let service = service();
    let scope = create_scope(&service).await;
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context()),
            scope_id: scope.id,
            policy_id: "policy_passive_dns".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    let mut events = service
        .watch_task(Request::new(WatchTaskRequest {
            context: Some(context()),
            task_id: task.id.clone(),
            after_sequence: 0,
        }))
        .await
        .unwrap()
        .into_inner();

    let queued = tokio::time::timeout(Duration::from_secs(1), events.next())
        .await
        .unwrap()
        .unwrap()
        .unwrap();
    assert_eq!(queued.event_type, "task.queued");

    service
        .cancel_task(Request::new(CancelTaskRequest {
            context: Some(context()),
            task_id: task.id,
        }))
        .await
        .unwrap();
    let canceled = tokio::time::timeout(Duration::from_secs(1), events.next())
        .await
        .unwrap()
        .unwrap()
        .unwrap();
    assert_eq!(canceled.event_type, "task.canceled");
}

#[tokio::test]
async fn denies_missing_capability() {
    struct Deny;

    impl Authorizer for Deny {
        fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
            false
        }
    }

    let service = CyberEdgeService::new(Arc::new(MemoryRepository::default()), Arc::new(Deny));
    let error = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context()),
            name: "Denied".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap_err();

    assert_eq!(error.code(), Code::PermissionDenied);
    assert_eq!(
        ErrorDetail::decode(error.details()).unwrap().code,
        "CAPABILITY_DENIED"
    );
}
