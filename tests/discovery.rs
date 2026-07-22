use std::{
    net::{IpAddr, Ipv4Addr},
    sync::Arc,
};

use async_trait::async_trait;
use cyberedge::{
    Authorizer, CyberEdgeService, DiscoveryWorker, DnsResolver, MemoryRepository, Repository,
    proto::{
        CreateScopeRequest, GetEvidenceRequest, GetTaskReportRequest, InvocationContext,
        ScopeTarget, SearchAssetsRequest, SearchAuditRequest, SearchObservationsRequest,
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

fn context(suffix: &str) -> InvocationContext {
    InvocationContext {
        request_id: format!("req_{suffix}"),
        idempotency_key: format!("idem_{suffix}"),
        agent_id: "agent_test".to_owned(),
        skill_name: "cyberedge-discover-assets".to_owned(),
        skill_version: "0.1.0".to_owned(),
    }
}
