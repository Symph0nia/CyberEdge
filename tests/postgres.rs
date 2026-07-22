use std::{
    net::{IpAddr, Ipv4Addr},
    sync::{
        Arc,
        atomic::{AtomicUsize, Ordering},
    },
};

use async_trait::async_trait;
use cyberedge::{
    Authorizer, CertificateProbe, CyberEdgeService, DiscoveryWorker, DnsResolver, PortConnector,
    PostgresRepository, Repository, WebSnapshot, WebsiteProbe,
    proto::{
        AssetChangeKind, CreateScheduleRequest, CreateScopeRequest, ExposureChangeKind,
        FindingSeverity, GetTaskReportRequest, GetTaskRequest, InvocationContext,
        ReportFindingRequest, ScopeTarget, SearchAssetChangesRequest, SearchAuditRequest,
        SearchCertificatesRequest, SearchExposureChangesRequest, SearchFindingsRequest,
        SearchSchedulesRequest, SearchServicesRequest, SearchWebsitesRequest, StartScanRequest,
        TargetKind, WatchTaskRequest, cyber_edge_server::CyberEdge,
    },
};
use sqlx::PgPool;
use tokio_stream::StreamExt;
use tonic::Request;

struct TestAuthorizer;

struct Resolver(AtomicUsize);

struct OpenConnector;
struct TestCertificate(Vec<u8>);
struct TestWebsite;
struct ChangingWebsite(AtomicUsize);

#[async_trait]
impl DnsResolver for Resolver {
    async fn resolve(&self, _domain: &str) -> Result<Vec<IpAddr>, String> {
        let suffix = 20 + self.0.fetch_add(1, Ordering::SeqCst) as u8;
        Ok(vec![IpAddr::V4(Ipv4Addr::new(192, 0, 2, suffix))])
    }
}

#[async_trait]
impl PortConnector for OpenConnector {
    async fn connect(&self, _address: IpAddr, port: u16) -> Result<bool, String> {
        Ok(port == 443)
    }
}

#[async_trait]
impl CertificateProbe for TestCertificate {
    async fn leaf_certificate(
        &self,
        _address: IpAddr,
        _port: u16,
        _server_name: &str,
    ) -> Result<Vec<u8>, String> {
        Ok(self.0.clone())
    }
}

#[async_trait]
impl WebsiteProbe for TestWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<WebSnapshot, String> {
        Ok(WebSnapshot {
            url: format!("https://{server_name}:{port}/"),
            status_code: 200,
            title: "Postgres Test".to_owned(),
            server: "test".to_owned(),
            content_type: "text/html".to_owned(),
            body: b"<title>Postgres Test</title>".to_vec(),
        })
    }
}

#[async_trait]
impl WebsiteProbe for ChangingWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<WebSnapshot, String> {
        let version = self.0.fetch_add(1, Ordering::SeqCst) + 1;
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: format!("Version {version}"),
            server: "test".to_owned(),
            content_type: "text/html".to_owned(),
            body: format!("version {version}").into_bytes(),
        })
    }
}

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

    let repository: Arc<dyn Repository> = Arc::new(repository);
    let service = CyberEdgeService::new(repository.clone(), authorizer());
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
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_dns".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("schedule")),
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_dns".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let scheduled_tasks = repository
        .enqueue_due_schedules(schedule.next_run_at.unwrap())
        .await
        .unwrap();
    assert_eq!(scheduled_tasks.len(), 1);
    let certificate = rcgen::generate_simple_self_signed(vec!["example.com".to_owned()]).unwrap();
    let worker = DiscoveryWorker::new(repository.clone(), Arc::new(Resolver(AtomicUsize::new(0))))
        .with_port_connector(Arc::new(OpenConnector), vec![443])
        .with_certificate_probe(Arc::new(TestCertificate(certificate.cert.der().to_vec())))
        .with_website_probe(Arc::new(TestWebsite));
    assert!(worker.run_once().await.unwrap());
    assert!(worker.run_once().await.unwrap());
    let next = repository.search_schedules(&scope.id).await.unwrap()[0]
        .next_run_at
        .unwrap();
    let latest_scheduled_tasks = repository.enqueue_due_schedules(next).await.unwrap();
    assert!(worker.run_once().await.unwrap());
    let changes = service
        .search_asset_changes(Request::new(SearchAssetChangesRequest {
            context: Some(context("changes")),
            schedule_id: schedule.id.clone(),
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
    let service_task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("active-services")),
            scope_id: scope.id.clone(),
            policy_id: "policy_service_baseline".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert!(worker.run_once().await.unwrap());
    let services = service
        .search_services(Request::new(SearchServicesRequest {
            context: Some(context("services")),
            scope_id: scope.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .services;
    assert_eq!(services.len(), 1);
    assert_eq!(services[0].port, 443);
    let certificates = service
        .search_certificates(Request::new(SearchCertificatesRequest {
            context: Some(context("certificates")),
            scope_id: scope.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .certificates;
    assert_eq!(certificates.len(), 1);
    assert_eq!(certificates[0].dns_names, ["example.com"]);
    let websites = service
        .search_websites(Request::new(SearchWebsitesRequest {
            context: Some(context("websites")),
            scope_id: scope.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .websites;
    assert_eq!(websites.len(), 1);
    assert_eq!(websites[0].title, "Postgres Test");
    let service_report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("service-report")),
            task_id: service_task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(service_report.services.len(), 1);
    assert_eq!(service_report.certificates.len(), 1);
    assert_eq!(service_report.websites.len(), 1);
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

    assert_eq!(stored.policy_id, "policy_passive_dns");
    assert_eq!(
        events.next().await.unwrap().unwrap().event_type,
        "task.queued"
    );
    assert_eq!(
        events.next().await.unwrap().unwrap().event_type,
        "task.running"
    );
    assert_eq!(
        events.next().await.unwrap().unwrap().event_type,
        "task.completed"
    );
    let audit_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM audit_events")
        .fetch_one(&pool)
        .await
        .unwrap();
    let outbox_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM outbox_events")
        .fetch_one(&pool)
        .await
        .unwrap();
    assert_eq!(audit_count, 4);
    let asset_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM assets")
        .fetch_one(&pool)
        .await
        .unwrap();
    let observation_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM observations")
        .fetch_one(&pool)
        .await
        .unwrap();
    assert_eq!(outbox_count, 16);
    assert_eq!(asset_count, 5);
    assert_eq!(observation_count, 10);
    let report = restarted
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("report")),
            task_id: stored.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.assets.len(), 2);
    assert_eq!(report.evidence.len(), 2);
    let audit = restarted
        .search_audit(Request::new(SearchAuditRequest {
            context: Some(context("audit")),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(audit.events.len(), 4);
    let schedules = restarted
        .search_schedules(Request::new(SearchSchedulesRequest {
            context: Some(context("schedules")),
            scope_id: scope.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .schedules;
    assert_eq!(schedules[0].last_task_id, latest_scheduled_tasks[0].id);

    let finding = restarted
        .report_finding(Request::new(ReportFindingRequest {
            context: Some(context("finding")),
            task_id: stored.id.clone(),
            observation_id: report.observations[0].id.clone(),
            detector: "postgres-test".to_owned(),
            rule_id: "evidence-chain".to_owned(),
            title: "Persistent finding".to_owned(),
            description: "Finding linked to persisted evidence".to_owned(),
            severity: FindingSeverity::Medium.into(),
            fingerprint: "postgres-finding-fingerprint".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(finding.evidence_id, report.observations[0].evidence_id);
    let findings = restarted
        .search_findings(Request::new(SearchFindingsRequest {
            context: Some(context("findings")),
            scope_id: scope.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .findings;
    assert_eq!(findings.len(), 1);

    let monitor_scope = restarted
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("exposure-scope")),
            name: "Exposure monitor".to_owned(),
            authorization_ref: "authorization:postgres-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let monitor_schedule = restarted
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("exposure-schedule")),
            scope_id: monitor_scope.id,
            policy_id: "policy_service_baseline".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let monitor_worker =
        DiscoveryWorker::new(repository.clone(), Arc::new(Resolver(AtomicUsize::new(0))))
            .with_port_connector(Arc::new(OpenConnector), vec![443])
            .with_website_probe(Arc::new(ChangingWebsite(AtomicUsize::new(0))));
    repository
        .enqueue_due_schedules(monitor_schedule.next_run_at.unwrap())
        .await
        .unwrap();
    monitor_worker.run_once().await.unwrap();
    let next = repository
        .search_schedules(&monitor_schedule.scope_id)
        .await
        .unwrap()[0]
        .next_run_at
        .unwrap();
    repository.enqueue_due_schedules(next).await.unwrap();
    monitor_worker.run_once().await.unwrap();
    let exposure_changes = restarted
        .search_exposure_changes(Request::new(SearchExposureChangesRequest {
            context: Some(context("exposure-changes")),
            schedule_id: monitor_schedule.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .changes;
    assert_eq!(exposure_changes.len(), 1);
    assert_eq!(
        exposure_changes[0].kind,
        i32::from(ExposureChangeKind::Modified)
    );
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
        "TRUNCATE findings, observations, evidence, websites, certificates, services, assets, outbox_events, audit_events,
         idempotency_keys, exposure_changes, asset_changes, task_events, tasks, schedules, scope_targets, scopes
         RESTART IDENTITY CASCADE",
    )
    .execute(pool)
    .await
    .unwrap();
}
