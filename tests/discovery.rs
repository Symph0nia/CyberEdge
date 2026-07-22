use std::{
    net::{IpAddr, Ipv4Addr},
    sync::{
        Arc,
        atomic::{AtomicUsize, Ordering},
    },
};

use async_trait::async_trait;
use cyberedge::{
    Authorizer, CertificateSource, CyberEdgeService, DiscoveryWorker, DnsResolver,
    MemoryRepository, Repository,
    proto::{
        AssetChangeKind, CreateScheduleRequest, CreateScopeRequest, GetEvidenceRequest,
        GetTaskReportRequest, InvocationContext, ScopeTarget, SearchAssetChangesRequest,
        SearchAssetsRequest, SearchAuditRequest, SearchObservationsRequest, SearchSchedulesRequest,
        StartScanRequest, TargetKind, TaskState, cyber_edge_server::CyberEdge,
    },
};
use sha2::{Digest, Sha256};
use tonic::Request;

struct Allow;

impl Authorizer for Allow {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

struct Resolver;

#[async_trait]
impl DnsResolver for Resolver {
    async fn resolve(&self, domain: &str) -> Result<Vec<IpAddr>, String> {
        assert_eq!(domain, "example.com");
        Ok(vec![IpAddr::V4(Ipv4Addr::new(192, 0, 2, 10))])
    }
}

struct Certificates;

#[async_trait]
impl CertificateSource for Certificates {
    async fn discover(&self, domain: &str) -> Result<Vec<String>, String> {
        assert_eq!(domain, "example.com");
        Ok(vec![
            "*.API.Example.com".to_owned(),
            "api.example.com.".to_owned(),
            "example.com".to_owned(),
            "notexample.com".to_owned(),
            "bad name.example.com".to_owned(),
        ])
    }
}

struct ChangingResolver(AtomicUsize);

#[async_trait]
impl DnsResolver for ChangingResolver {
    async fn resolve(&self, _domain: &str) -> Result<Vec<IpAddr>, String> {
        let suffix = if self.0.fetch_add(1, Ordering::SeqCst) == 0 {
            10
        } else {
            11
        };
        Ok(vec![IpAddr::V4(Ipv4Addr::new(192, 0, 2, suffix))])
    }
}

struct FailingResolver(AtomicUsize);

#[async_trait]
impl DnsResolver for FailingResolver {
    async fn resolve(&self, _domain: &str) -> Result<Vec<IpAddr>, String> {
        if self.0.fetch_add(1, Ordering::SeqCst) == 0 {
            Ok(vec![IpAddr::V4(Ipv4Addr::new(192, 0, 2, 20))])
        } else {
            Err("upstream timeout".to_owned())
        }
    }
}

#[tokio::test]
async fn executes_passive_dns_and_retains_evidence() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("scope")),
            name: "Discovery".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("start")),
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_dns".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();

    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver));
    assert!(worker.run_once().await.unwrap());
    assert!(!worker.run_once().await.unwrap());

    let stored = service
        .get_task(Request::new(cyberedge::proto::GetTaskRequest {
            context: Some(context("get-task")),
            task_id: task.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(stored.state, i32::from(TaskState::Completed));

    let assets = service
        .search_assets(Request::new(SearchAssetsRequest {
            context: Some(context("assets")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .assets;
    assert_eq!(assets.len(), 2);
    assert!(assets.iter().any(|asset| asset.value == "example.com"));
    assert!(assets.iter().any(|asset| asset.value == "192.0.2.10"));

    let observations = service
        .search_observations(Request::new(SearchObservationsRequest {
            context: Some(context("observations")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .observations;
    assert_eq!(observations.len(), 2);
    let evidence = service
        .get_evidence(Request::new(GetEvidenceRequest {
            context: Some(context("evidence")),
            evidence_id: observations[0].evidence_id.clone(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(
        format!("{:x}", Sha256::digest(&evidence.content)),
        evidence.sha256
    );
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("report")),
            task_id: stored.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.assets.len(), 2);
    assert_eq!(report.observations.len(), 2);
    assert_eq!(report.evidence.len(), 2);
    let audit = service
        .search_audit(Request::new(SearchAuditRequest {
            context: Some(context("audit")),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(audit.events.len(), 2);
}

#[tokio::test]
async fn certificate_inventory_keeps_only_normalized_in_scope_domains() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("ct-scope")),
            name: "CT Discovery".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("ct-start")),
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_inventory".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();

    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_certificate_source(Arc::new(Certificates));
    assert!(worker.run_once().await.unwrap());

    let assets = service
        .search_assets(Request::new(SearchAssetsRequest {
            context: Some(context("ct-assets")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .assets;
    let values = assets
        .into_iter()
        .map(|asset| asset.value)
        .collect::<std::collections::BTreeSet<_>>();
    assert_eq!(
        values,
        ["192.0.2.10", "api.example.com", "example.com"]
            .into_iter()
            .map(str::to_owned)
            .collect()
    );

    let observations = service
        .search_observations(Request::new(SearchObservationsRequest {
            context: Some(context("ct-observations")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .observations;
    assert_eq!(
        observations
            .iter()
            .filter(|item| item.observation_type == "ct.discovered_domain")
            .count(),
        2
    );
}

#[tokio::test]
async fn due_schedule_creates_a_normal_task() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("schedule-scope")),
            name: "Recurring discovery".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("schedule-create")),
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_dns".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let due = schedule.next_run_at.unwrap();
    let tasks = repository.enqueue_due_schedules(due).await.unwrap();
    assert_eq!(tasks.len(), 1);
    assert_eq!(tasks[0].scope_id, scope.id);
    assert_eq!(tasks[0].state, i32::from(TaskState::Queued));

    let stored = service
        .search_schedules(Request::new(SearchSchedulesRequest {
            context: Some(context("schedule-search")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .schedules;
    assert_eq!(stored[0].last_task_id, tasks[0].id);

    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver));
    assert!(worker.run_once().await.unwrap());
    let task = service
        .get_task(Request::new(cyberedge::proto::GetTaskRequest {
            context: Some(context("schedule-task")),
            task_id: tasks[0].id.clone(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(task.state, i32::from(TaskState::Completed));
}

#[tokio::test]
async fn monitoring_records_appeared_and_disappeared_assets_after_baseline() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("monitor-scope")),
            name: "Asset monitor".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("monitor-schedule")),
            scope_id: scope.id,
            policy_id: "policy_passive_dns".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(
        repository.clone(),
        Arc::new(ChangingResolver(AtomicUsize::new(0))),
    );

    repository
        .enqueue_due_schedules(schedule.next_run_at.unwrap())
        .await
        .unwrap();
    worker.run_once().await.unwrap();
    assert!(
        repository
            .search_asset_changes(&schedule.id)
            .await
            .unwrap()
            .is_empty()
    );

    let next = repository
        .search_schedules(&schedule.scope_id)
        .await
        .unwrap()[0]
        .next_run_at
        .unwrap();
    repository.enqueue_due_schedules(next).await.unwrap();
    worker.run_once().await.unwrap();
    let changes = service
        .search_asset_changes(Request::new(SearchAssetChangesRequest {
            context: Some(context("monitor-changes")),
            schedule_id: schedule.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .changes;
    assert_eq!(changes.len(), 2);
    assert!(
        changes
            .iter()
            .any(|change| change.kind == i32::from(AssetChangeKind::Appeared))
    );
    assert!(
        changes
            .iter()
            .any(|change| change.kind == i32::from(AssetChangeKind::Disappeared))
    );
}

#[tokio::test]
async fn monitoring_does_not_report_disappearance_when_coverage_fails() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("coverage-scope")),
            name: "Coverage guard".to_owned(),
            authorization_ref: "authorization:test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "example.com".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("coverage-schedule")),
            scope_id: scope.id,
            policy_id: "policy_passive_dns".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(
        repository.clone(),
        Arc::new(FailingResolver(AtomicUsize::new(0))),
    );
    let first = schedule.next_run_at.unwrap();
    repository.enqueue_due_schedules(first).await.unwrap();
    worker.run_once().await.unwrap();
    let second = repository
        .search_schedules(&schedule.scope_id)
        .await
        .unwrap()[0]
        .next_run_at
        .unwrap();
    repository.enqueue_due_schedules(second).await.unwrap();
    worker.run_once().await.unwrap();

    assert!(
        repository
            .search_asset_changes(&schedule.id)
            .await
            .unwrap()
            .is_empty()
    );
}

fn context(suffix: &str) -> InvocationContext {
    InvocationContext {
        request_id: format!("req_{suffix}"),
        idempotency_key: format!("idem_{suffix}"),
        agent_id: "agent_test".to_owned(),
        skill_name: "cyberedge-discover-assets".to_owned(),
        skill_version: "0.1.0".to_owned(),
    }
}
