use std::{
    net::{IpAddr, Ipv4Addr},
    sync::{
        Arc,
        atomic::{AtomicUsize, Ordering},
    },
};

use async_trait::async_trait;
use base64::{Engine, engine::general_purpose::STANDARD};
use cyberedge::{
    Authorizer, CertificateProbe, CertificateSource, CyberEdgeService, DiscoveryWorker,
    DnsResolver, MemoryRepository, NucleiProbe, PortConnector, Repository, ScreenshotProbe,
    SystemCertificateProbe, SystemPortConnector, SystemWebsiteProbe, WebSnapshot, WebsiteProbe,
    proto::{
        AssetChangeKind, CreateScheduleRequest, CreateScopeRequest, ExposureChangeKind,
        FindingSeverity, FindingState, GetEvidenceRequest, GetTaskReportRequest, InvocationContext,
        ReportFindingRequest, ScopeTarget, SearchAssetChangesRequest, SearchAssetsRequest,
        SearchAuditRequest, SearchCertificatesRequest, SearchExposureChangesRequest,
        SearchFindingsRequest, SearchObservationsRequest, SearchSchedulesRequest,
        SearchServicesRequest, SearchWebsitesRequest, StartScanRequest, TargetKind, TaskState,
        cyber_edge_server::CyberEdge,
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

struct OpenPort;

#[async_trait]
impl PortConnector for OpenPort {
    async fn connect(&self, _address: IpAddr, _port: u16) -> Result<bool, String> {
        Ok(true)
    }
}

struct TestCertificate(Vec<u8>);
struct CyclingCertificate {
    expired: Vec<u8>,
    healthy: Vec<u8>,
    calls: AtomicUsize,
}
struct TestWebsite;
struct TestScreenshot;
struct DirectoryWebsite;
struct CyclingDirectoryWebsite(AtomicUsize);
struct CyclingExposureWebsite(AtomicUsize);
struct BoundedCrawlerWebsite;
struct ChangingWebsite(AtomicUsize);
struct FailingWebsite(AtomicUsize);
struct HostCollisionResolver;
struct CyclingHostCollisionWebsite(AtomicUsize);
struct CyclingNuclei(AtomicUsize);

#[async_trait]
impl DnsResolver for HostCollisionResolver {
    async fn resolve(&self, domain: &str) -> Result<Vec<IpAddr>, String> {
        assert_eq!(domain, "example.com");
        Ok(vec![IpAddr::V4(Ipv4Addr::new(192, 0, 2, 10))])
    }
}

#[async_trait]
impl WebsiteProbe for CyclingHostCollisionWebsite {
    async fn fetch(
        &self,
        address: IpAddr,
        port: u16,
        server_name: &str,
        path: &str,
    ) -> Result<WebSnapshot, String> {
        assert_eq!(path, "/");
        let collision_candidate =
            address == IpAddr::V4(Ipv4Addr::new(192, 0, 2, 20)) && server_name == "example.com";
        let exposed = collision_candidate && self.0.fetch_add(1, Ordering::SeqCst) != 1;
        let (title, body) = if exposed {
            (
                "Private Portal",
                format!(
                    "<html><title>Private Portal</title><body>{}</body></html>",
                    "authorized host route ".repeat(10)
                ),
            )
        } else {
            (
                "Default Site",
                format!(
                    "<html><title>Default Site</title><body>{}</body></html>",
                    "default address response ".repeat(10)
                ),
            )
        };
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: title.to_owned(),
            server: "test".to_owned(),
            content_type: "text/html".to_owned(),
            body: body.into_bytes(),
        })
    }
}

#[async_trait]
impl NucleiProbe for CyclingNuclei {
    async fn scan(&self, targets: &[String]) -> Result<Vec<u8>, String> {
        assert_eq!(targets.len(), 1);
        if self.0.fetch_add(1, Ordering::SeqCst) == 1 {
            return Ok(Vec::new());
        }
        Ok(serde_json::to_vec(&serde_json::json!({
            "template-id": "CVE-2099-0001",
            "matcher-name": "version-check",
            "type": "http",
            "host": targets[0],
            "matched-at": format!("{}admin", targets[0]),
            "matcher-status": true,
            "curl-command": "curl http://discarded.example/",
            "info": {
                "name": "Test application vulnerability",
                "description": "The retained scanner result matched the reviewed template.",
                "severity": "high"
            }
        }))
        .unwrap())
    }
}

#[async_trait]
impl CertificateProbe for TestCertificate {
    async fn leaf_certificate(
        &self,
        _address: IpAddr,
        _port: u16,
        server_name: &str,
    ) -> Result<Vec<u8>, String> {
        assert_eq!(server_name, "127.0.0.1");
        Ok(self.0.clone())
    }
}

#[async_trait]
impl CertificateProbe for CyclingCertificate {
    async fn leaf_certificate(
        &self,
        _address: IpAddr,
        _port: u16,
        _server_name: &str,
    ) -> Result<Vec<u8>, String> {
        if self.calls.fetch_add(1, Ordering::SeqCst) == 1 {
            Ok(self.healthy.clone())
        } else {
            Ok(self.expired.clone())
        }
    }
}

#[async_trait]
impl WebsiteProbe for TestWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        _server_name: &str,
        _path: &str,
    ) -> Result<WebSnapshot, String> {
        Ok(WebSnapshot {
            url: format!("https://127.0.0.1:{port}/"),
            status_code: 200,
            title: "CyberEdge Test".to_owned(),
            server: "test-server".to_owned(),
            content_type: "text/html".to_owned(),
            body:
                b"<title>CyberEdge Test</title><meta name=\"generator\" content=\"WordPress 6.8\">"
                    .to_vec(),
        })
    }
}

#[async_trait]
impl ScreenshotProbe for TestScreenshot {
    async fn capture(&self, html: &[u8]) -> Result<Vec<u8>, String> {
        assert!(String::from_utf8_lossy(html).contains("CyberEdge Test"));
        Ok(b"\x89PNG\r\n\x1a\ncyberedge-test".to_vec())
    }
}

#[async_trait]
impl WebsiteProbe for DirectoryWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
        _path: &str,
    ) -> Result<WebSnapshot, String> {
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: "Index of /".to_owned(),
            server: "nginx".to_owned(),
            content_type: "text/html".to_owned(),
            body: b"<html><title>Index of /</title><a href=\"../\">Parent Directory</a></html>"
                .to_vec(),
        })
    }
}

#[async_trait]
impl WebsiteProbe for CyclingDirectoryWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
        _path: &str,
    ) -> Result<WebSnapshot, String> {
        let exposed = self.0.fetch_add(1, Ordering::SeqCst) != 1;
        let (title, body) = if exposed {
            (
                "Index of /",
                b"<title>Index of /</title><a href=\"../\">Parent Directory</a>".as_slice(),
            )
        } else {
            ("Application", b"<title>Application</title>".as_slice())
        };
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: title.to_owned(),
            server: "nginx".to_owned(),
            content_type: "text/html".to_owned(),
            body: body.to_vec(),
        })
    }
}

#[async_trait]
impl WebsiteProbe for CyclingExposureWebsite {
    fn supports_path_probes(&self) -> bool {
        true
    }

    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
        path: &str,
    ) -> Result<WebSnapshot, String> {
        let request = self.0.fetch_add(1, Ordering::SeqCst);
        let exposed = request / 3 != 1;
        let (status_code, content_type, body) = match (path, exposed) {
            ("/", _) => (200, "text/html", b"<title>Application</title>".to_vec()),
            ("/.git/HEAD", true) => (200, "text/plain", b"ref: refs/heads/main\n".to_vec()),
            ("/.DS_Store", true) => (
                200,
                "application/octet-stream",
                b"\0\0\0\x01Bud1test".to_vec(),
            ),
            (_, false) => (404, "text/plain", b"not found".to_vec()),
            _ => panic!("unexpected path {path}"),
        };
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}{path}"),
            status_code,
            title: String::new(),
            server: "test".to_owned(),
            content_type: content_type.to_owned(),
            body,
        })
    }
}

#[async_trait]
impl WebsiteProbe for BoundedCrawlerWebsite {
    fn supports_crawl(&self) -> bool {
        true
    }

    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
        path: &str,
    ) -> Result<WebSnapshot, String> {
        let body = match path {
            "/" => b"<title>Root</title><a href=\"/login\">Login</a><a href='/docs/start'>Docs</a><a href=\"https://outside.example/\">Outside</a><a href=\"/search?q=x\">Query</a><a href=\"/../admin\">Traversal</a>".to_vec(),
            "/login" => b"login evidence".to_vec(),
            "/docs/start" => b"documentation evidence".to_vec(),
            _ => panic!("crawler fetched unexpected path {path}"),
        };
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}{path}"),
            status_code: 200,
            title: if path == "/" { "Root" } else { "" }.to_owned(),
            server: "test".to_owned(),
            content_type: "text/html".to_owned(),
            body,
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
        _path: &str,
    ) -> Result<WebSnapshot, String> {
        let version = self.0.fetch_add(1, Ordering::SeqCst) + 1;
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: format!("Version {version}"),
            server: "test".to_owned(),
            content_type: "text/html".to_owned(),
            body: format!("<title>Version {version}</title>").into_bytes(),
        })
    }
}

#[async_trait]
impl WebsiteProbe for FailingWebsite {
    async fn fetch(
        &self,
        _address: IpAddr,
        port: u16,
        server_name: &str,
        _path: &str,
    ) -> Result<WebSnapshot, String> {
        if self.0.fetch_add(1, Ordering::SeqCst) > 0 {
            return Err("HTTP timeout".to_owned());
        }
        Ok(WebSnapshot {
            url: format!("http://{server_name}:{port}/"),
            status_code: 200,
            title: "Available".to_owned(),
            server: String::new(),
            content_type: "text/html".to_owned(),
            body: b"available".to_vec(),
        })
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
async fn finding_requires_task_observation_and_deduplicates_fingerprint() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("finding-scope")),
            name: "Finding scope".to_owned(),
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
            context: Some(context("finding-task")),
            scope_id: scope.id.clone(),
            policy_id: "policy_passive_dns".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    DiscoveryWorker::new(repository, Arc::new(Resolver))
        .run_once()
        .await
        .unwrap();
    let observation = service
        .search_observations(Request::new(SearchObservationsRequest {
            context: Some(context("finding-observations")),
            task_id: task.id.clone(),
        }))
        .await
        .unwrap()
        .into_inner()
        .observations
        .into_iter()
        .next()
        .unwrap();
    let request = |suffix: &str| ReportFindingRequest {
        context: Some(context(suffix)),
        task_id: task.id.clone(),
        observation_id: observation.id.clone(),
        detector: "test-detector".to_owned(),
        rule_id: "dns-observation-review".to_owned(),
        title: "Observed DNS condition".to_owned(),
        description: "Evidence-backed test finding".to_owned(),
        severity: FindingSeverity::Low.into(),
        fingerprint: "sha256:test-fingerprint".to_owned(),
    };
    let first = service
        .report_finding(Request::new(request("finding-report-1")))
        .await
        .unwrap()
        .into_inner();
    let repeated = service
        .report_finding(Request::new(request("finding-report-2")))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(first.id, repeated.id);
    assert_eq!(first.evidence_id, observation.evidence_id);
    let findings = service
        .search_findings(Request::new(SearchFindingsRequest {
            context: Some(context("finding-search")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .findings;
    assert_eq!(findings.len(), 1);
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("finding-report")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.findings.len(), 1);
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

#[tokio::test]
async fn monitoring_records_modified_website_without_reopening_service() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("website-monitor-scope")),
            name: "Website monitor".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("website-monitor-schedule")),
            scope_id: scope.id,
            policy_id: "policy_service_baseline".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository.clone(), Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(ChangingWebsite(AtomicUsize::new(0))));
    repository
        .enqueue_due_schedules(schedule.next_run_at.unwrap())
        .await
        .unwrap();
    worker.run_once().await.unwrap();
    assert!(
        repository
            .search_exposure_changes(&schedule.id)
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
        .search_exposure_changes(Request::new(SearchExposureChangesRequest {
            context: Some(context("website-monitor-changes")),
            schedule_id: schedule.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .changes;
    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].resource_kind, "website");
    assert_eq!(changes[0].kind, i32::from(ExposureChangeKind::Modified));
    assert!(!changes[0].previous_fingerprint.is_empty());
    assert!(!changes[0].current_fingerprint.is_empty());
}

#[tokio::test]
async fn monitoring_does_not_report_website_disappeared_after_http_error() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("website-error-scope")),
            name: "Website error coverage".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let schedule = service
        .create_schedule(Request::new(CreateScheduleRequest {
            context: Some(context("website-error-schedule")),
            scope_id: scope.id,
            policy_id: "policy_service_baseline".to_owned(),
            interval_seconds: 60,
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository.clone(), Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(FailingWebsite(AtomicUsize::new(0))));
    repository
        .enqueue_due_schedules(schedule.next_run_at.unwrap())
        .await
        .unwrap();
    worker.run_once().await.unwrap();
    let next = repository
        .search_schedules(&schedule.scope_id)
        .await
        .unwrap()[0]
        .next_run_at
        .unwrap();
    repository.enqueue_due_schedules(next).await.unwrap();
    worker.run_once().await.unwrap();

    assert!(
        service
            .search_exposure_changes(Request::new(SearchExposureChangesRequest {
                context: Some(context("website-error-changes")),
                schedule_id: schedule.id,
            }))
            .await
            .unwrap()
            .into_inner()
            .changes
            .is_empty()
    );
}

#[tokio::test]
async fn discovers_service_on_authorized_local_listener() {
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let port = listener.local_addr().unwrap().port();
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("service-scope")),
            name: "Local service".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("service-start")),
            scope_id: scope.id.clone(),
            policy_id: "policy_service_baseline".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();

    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(SystemPortConnector), vec![port]);
    assert!(worker.run_once().await.unwrap());
    let services = service
        .search_services(Request::new(SearchServicesRequest {
            context: Some(context("services")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .services;
    assert_eq!(services.len(), 1);
    assert_eq!(services[0].port, u32::from(port));
    assert_eq!(services[0].transport, "tcp");
    assert_eq!(services[0].service_hint, "unknown");
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("service-report")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.services.len(), 1);
    assert_eq!(report.services[0].port, u32::from(port));
    drop(listener);
}

#[tokio::test]
async fn reports_directory_listing_from_retained_http_evidence() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("directory-scope")),
            name: "Directory listing".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("directory-start")),
            scope_id: scope.id.clone(),
            policy_id: "policy_service_baseline".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(DirectoryWebsite))
        .run_once()
        .await
        .unwrap();

    let findings = service
        .search_findings(Request::new(SearchFindingsRequest {
            context: Some(context("directory-findings")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .findings;
    assert_eq!(findings.len(), 1);
    assert_eq!(findings[0].rule_id, "http-directory-listing-v1");
    assert_eq!(findings[0].severity, i32::from(FindingSeverity::Medium));
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("directory-report")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.findings.len(), 1);
    let evidence = report
        .evidence
        .iter()
        .find(|evidence| evidence.id == findings[0].evidence_id)
        .unwrap();
    assert!(String::from_utf8_lossy(&evidence.content).contains("Parent Directory"));
}

#[tokio::test]
async fn resolves_and_reopens_builtin_finding_after_successful_reevaluation() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("finding-lifecycle-scope")),
            name: "Finding lifecycle".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(CyclingDirectoryWebsite(AtomicUsize::new(0))));

    let mut finding_id = String::new();
    for (index, expected_state) in [
        FindingState::Open,
        FindingState::Resolved,
        FindingState::Open,
    ]
    .into_iter()
    .enumerate()
    {
        service
            .start_scan(Request::new(StartScanRequest {
                context: Some(context(&format!("finding-lifecycle-task-{index}"))),
                scope_id: scope.id.clone(),
                policy_id: "policy_service_baseline".to_owned(),
            }))
            .await
            .unwrap();
        worker.run_once().await.unwrap();
        let findings = service
            .search_findings(Request::new(SearchFindingsRequest {
                context: Some(context(&format!("finding-lifecycle-search-{index}"))),
                scope_id: scope.id.clone(),
            }))
            .await
            .unwrap()
            .into_inner()
            .findings;
        assert_eq!(findings.len(), 1);
        if finding_id.is_empty() {
            finding_id = findings[0].id.clone();
        }
        assert_eq!(findings[0].id, finding_id);
        assert_eq!(findings[0].state, i32::from(expected_state));
    }
}

#[tokio::test]
async fn detects_resolves_and_reopens_fixed_http_exposure_probes() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("exposure-probe-scope")),
            name: "HTTP exposure probes".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(CyclingExposureWebsite(AtomicUsize::new(0))));
    let mut finding_ids = std::collections::BTreeSet::new();

    for (index, expected_state) in [
        FindingState::Open,
        FindingState::Resolved,
        FindingState::Open,
    ]
    .into_iter()
    .enumerate()
    {
        let task = service
            .start_scan(Request::new(StartScanRequest {
                context: Some(context(&format!("exposure-probe-task-{index}"))),
                scope_id: scope.id.clone(),
                policy_id: "policy_service_baseline".to_owned(),
            }))
            .await
            .unwrap()
            .into_inner();
        worker.run_once().await.unwrap();
        let findings = service
            .search_findings(Request::new(SearchFindingsRequest {
                context: Some(context(&format!("exposure-probe-search-{index}"))),
                scope_id: scope.id.clone(),
            }))
            .await
            .unwrap()
            .into_inner()
            .findings;
        assert_eq!(findings.len(), 2);
        assert!(
            findings
                .iter()
                .all(|finding| finding.state == i32::from(expected_state))
        );
        let current_ids = findings
            .iter()
            .map(|finding| finding.id.clone())
            .collect::<std::collections::BTreeSet<_>>();
        if finding_ids.is_empty() {
            finding_ids = current_ids.clone();
        }
        assert_eq!(current_ids, finding_ids);
        if index == 0 {
            let report = service
                .get_task_report(Request::new(GetTaskReportRequest {
                    context: Some(context("exposure-probe-report")),
                    task_id: task.id,
                }))
                .await
                .unwrap()
                .into_inner();
            assert_eq!(report.findings.len(), 2);
            assert!(
                report
                    .evidence
                    .iter()
                    .any(|evidence| { evidence.content.starts_with(b"ref: refs/heads/") })
            );
            assert!(
                report
                    .evidence
                    .iter()
                    .any(|evidence| { evidence.content.starts_with(b"\0\0\0\x01Bud1") })
            );
        }
    }
}

#[tokio::test]
async fn crawls_only_bounded_same_origin_paths_and_retains_evidence() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("crawler-scope")),
            name: "Bounded crawler".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("crawler-task")),
            scope_id: scope.id.clone(),
            policy_id: "policy_service_baseline".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(BoundedCrawlerWebsite))
        .run_once()
        .await
        .unwrap();

    let websites = service
        .search_websites(Request::new(SearchWebsitesRequest {
            context: Some(context("crawler-websites")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .websites;
    assert_eq!(websites[0].discovered_paths, ["/docs/start", "/login"]);
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("crawler-report")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(
        report
            .observations
            .iter()
            .filter(|observation| observation.observation_type == "http.crawl")
            .count(),
        2
    );
    assert!(
        report
            .evidence
            .iter()
            .any(|evidence| evidence.content == b"login evidence")
    );
    assert!(
        report
            .evidence
            .iter()
            .any(|evidence| evidence.content == b"documentation evidence")
    );
}

#[tokio::test]
async fn detects_resolves_and_reopens_host_collision_with_comparison_evidence() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("host-collision-scope")),
            name: "Host collision lifecycle".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![
                ScopeTarget {
                    kind: TargetKind::Domain.into(),
                    value: "example.com".to_owned(),
                },
                ScopeTarget {
                    kind: TargetKind::Ip.into(),
                    value: "192.0.2.20".to_owned(),
                },
            ],
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(HostCollisionResolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(CyclingHostCollisionWebsite(AtomicUsize::new(0))));

    let mut finding_id = String::new();
    for (index, expected_state) in [
        FindingState::Open,
        FindingState::Resolved,
        FindingState::Open,
    ]
    .into_iter()
    .enumerate()
    {
        let task = service
            .start_scan(Request::new(StartScanRequest {
                context: Some(context(&format!("host-collision-task-{index}"))),
                scope_id: scope.id.clone(),
                policy_id: "policy_service_baseline".to_owned(),
            }))
            .await
            .unwrap()
            .into_inner();
        worker.run_once().await.unwrap();
        let findings = service
            .search_findings(Request::new(SearchFindingsRequest {
                context: Some(context(&format!("host-collision-findings-{index}"))),
                scope_id: scope.id.clone(),
            }))
            .await
            .unwrap()
            .into_inner()
            .findings;
        assert_eq!(findings.len(), 1);
        assert_eq!(findings[0].state, i32::from(expected_state));
        assert_eq!(findings[0].rule_id, "http-host-collision-v1");
        if finding_id.is_empty() {
            finding_id = findings[0].id.clone();
        } else {
            assert_eq!(findings[0].id, finding_id);
        }

        let report = service
            .get_task_report(Request::new(GetTaskReportRequest {
                context: Some(context(&format!("host-collision-report-{index}"))),
                task_id: task.id,
            }))
            .await
            .unwrap()
            .into_inner();
        let collision_observations = report
            .observations
            .iter()
            .filter(|item| item.observation_type == "http.host_collision_check")
            .collect::<Vec<_>>();
        assert_eq!(collision_observations.len(), 1);
        let observation = collision_observations
            .first()
            .expect("host collision comparison observation");
        let evidence = report
            .evidence
            .iter()
            .find(|item| item.id == observation.evidence_id)
            .expect("host collision comparison evidence");
        let evidence: serde_json::Value = serde_json::from_slice(&evidence.content).unwrap();
        let baseline = STANDARD
            .decode(evidence["baseline"]["body_base64"].as_str().unwrap())
            .unwrap();
        assert!(String::from_utf8_lossy(&baseline).contains("default address response"));
        if expected_state == FindingState::Open {
            let candidate = STANDARD
                .decode(evidence["candidate"]["body_base64"].as_str().unwrap())
                .unwrap();
            assert!(String::from_utf8_lossy(&candidate).contains("authorized host route"));
        }
    }
}

#[tokio::test]
async fn nuclei_policy_retains_results_and_resolves_missing_templates() {
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("nuclei-scope")),
            name: "Nuclei lifecycle".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![80])
        .with_website_probe(Arc::new(TestWebsite))
        .with_nuclei_probe(Arc::new(CyclingNuclei(AtomicUsize::new(0))));

    let mut finding_id = String::new();
    for (index, expected_state) in [
        FindingState::Open,
        FindingState::Resolved,
        FindingState::Open,
    ]
    .into_iter()
    .enumerate()
    {
        let task = service
            .start_scan(Request::new(StartScanRequest {
                context: Some(context(&format!("nuclei-task-{index}"))),
                scope_id: scope.id.clone(),
                policy_id: "policy_vulnerability_baseline".to_owned(),
            }))
            .await
            .unwrap()
            .into_inner();
        worker.run_once().await.unwrap();
        let findings = service
            .search_findings(Request::new(SearchFindingsRequest {
                context: Some(context(&format!("nuclei-findings-{index}"))),
                scope_id: scope.id.clone(),
            }))
            .await
            .unwrap()
            .into_inner()
            .findings;
        assert_eq!(findings.len(), 1);
        assert_eq!(findings[0].state, i32::from(expected_state));
        assert_eq!(findings[0].detector, "nuclei");
        assert_eq!(findings[0].rule_id, "CVE-2099-0001");
        if finding_id.is_empty() {
            finding_id = findings[0].id.clone();
        } else {
            assert_eq!(findings[0].id, finding_id);
        }

        let report = service
            .get_task_report(Request::new(GetTaskReportRequest {
                context: Some(context(&format!("nuclei-report-{index}"))),
                task_id: task.id,
            }))
            .await
            .unwrap()
            .into_inner();
        assert!(
            report
                .observations
                .iter()
                .any(|item| item.observation_type == "nuclei.coverage")
        );
        if expected_state == FindingState::Open {
            let observation = report
                .observations
                .iter()
                .find(|item| item.observation_type == "nuclei.result")
                .expect("nuclei result observation");
            let evidence = report
                .evidence
                .iter()
                .find(|item| item.id == observation.evidence_id)
                .expect("nuclei result evidence");
            let evidence: serde_json::Value = serde_json::from_slice(&evidence.content).unwrap();
            assert_eq!(evidence["template-id"], "CVE-2099-0001");
            assert!(evidence.get("curl-command").is_none());
            assert_eq!(
                evidence["cyberedge-source-sha256"]
                    .as_str()
                    .expect("source hash")
                    .len(),
                64
            );
        }
    }
}

#[tokio::test]
async fn retains_tls_certificate_as_inventory_and_der_evidence() {
    let certificate = rcgen::generate_simple_self_signed(vec!["localhost".to_owned()]).unwrap();
    let der = certificate.cert.der().to_vec();
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("tls-scope")),
            name: "Local TLS".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let task = service
        .start_scan(Request::new(StartScanRequest {
            context: Some(context("tls-start")),
            scope_id: scope.id.clone(),
            policy_id: "policy_service_baseline".to_owned(),
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![443])
        .with_certificate_probe(Arc::new(TestCertificate(der.clone())))
        .with_website_probe(Arc::new(TestWebsite))
        .with_screenshot_probe(Arc::new(TestScreenshot));
    assert!(worker.run_once().await.unwrap());

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
    assert_eq!(certificates[0].dns_names, ["localhost"]);
    assert_eq!(
        certificates[0].sha256,
        format!("{:x}", Sha256::digest(&der))
    );
    let websites = service
        .search_websites(Request::new(SearchWebsitesRequest {
            context: Some(context("websites")),
            scope_id: scope.id,
        }))
        .await
        .unwrap()
        .into_inner()
        .websites;
    assert_eq!(websites.len(), 1);
    assert_eq!(websites[0].title, "CyberEdge Test");
    assert_eq!(websites[0].fingerprints.len(), 1);
    assert_eq!(websites[0].fingerprints[0].name, "WordPress");
    assert_eq!(websites[0].fingerprints[0].version, "6.8");
    assert!(!websites[0].screenshot_evidence_id.is_empty());
    let report = service
        .get_task_report(Request::new(GetTaskReportRequest {
            context: Some(context("tls-report")),
            task_id: task.id,
        }))
        .await
        .unwrap()
        .into_inner();
    assert_eq!(report.certificates.len(), 1);
    assert_eq!(report.websites.len(), 1);
    assert!(report.findings.is_empty());
    assert_eq!(
        report.websites[0].fingerprints[0].evidence_id,
        format!("evidence_{}", report.websites[0].content_sha256)
    );
    let screenshot = report
        .evidence
        .iter()
        .find(|evidence| evidence.id == report.websites[0].screenshot_evidence_id)
        .unwrap();
    assert_eq!(screenshot.media_type, "image/png");
    assert!(screenshot.content.starts_with(b"\x89PNG\r\n\x1a\n"));
    let evidence = report
        .evidence
        .iter()
        .find(|item| item.media_type == "application/pkix-cert")
        .unwrap();
    assert_eq!(evidence.content, der);
}

#[tokio::test]
async fn resolves_and_reopens_expired_certificate_finding_after_replacement() {
    let certificate = |not_after_year| {
        let mut params = rcgen::CertificateParams::new(vec!["localhost".to_owned()]).unwrap();
        params.not_before = rcgen::date_time_ymd(2019, 1, 1);
        params.not_after = rcgen::date_time_ymd(not_after_year, 1, 1);
        let key = rcgen::KeyPair::generate().unwrap();
        params.self_signed(&key).unwrap().der().to_vec()
    };
    let repository: Arc<dyn Repository> = Arc::new(MemoryRepository::default());
    let service = CyberEdgeService::new(repository.clone(), Arc::new(Allow));
    let scope = service
        .create_scope(Request::new(CreateScopeRequest {
            context: Some(context("certificate-lifecycle-scope")),
            name: "Certificate lifecycle".to_owned(),
            authorization_ref: "authorization:local-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Ip.into(),
                value: "127.0.0.1".to_owned(),
            }],
        }))
        .await
        .unwrap()
        .into_inner();
    let worker = DiscoveryWorker::new(repository, Arc::new(Resolver))
        .with_port_connector(Arc::new(OpenPort), vec![443])
        .with_certificate_probe(Arc::new(CyclingCertificate {
            expired: certificate(2020),
            healthy: certificate(4090),
            calls: AtomicUsize::new(0),
        }))
        .with_website_probe(Arc::new(TestWebsite));

    let mut finding_id = String::new();
    for (index, expected_state) in [
        FindingState::Open,
        FindingState::Resolved,
        FindingState::Open,
    ]
    .into_iter()
    .enumerate()
    {
        service
            .start_scan(Request::new(StartScanRequest {
                context: Some(context(&format!("certificate-lifecycle-task-{index}"))),
                scope_id: scope.id.clone(),
                policy_id: "policy_service_baseline".to_owned(),
            }))
            .await
            .unwrap();
        worker.run_once().await.unwrap();
        let findings = service
            .search_findings(Request::new(SearchFindingsRequest {
                context: Some(context(&format!("certificate-lifecycle-search-{index}"))),
                scope_id: scope.id.clone(),
            }))
            .await
            .unwrap()
            .into_inner()
            .findings;
        assert_eq!(findings.len(), 1);
        assert_eq!(findings[0].rule_id, "tls-certificate-expired-v1");
        if finding_id.is_empty() {
            finding_id = findings[0].id.clone();
        }
        assert_eq!(findings[0].id, finding_id);
        assert_eq!(findings[0].state, i32::from(expected_state));
    }
}

#[tokio::test]
async fn system_certificate_probe_handshakes_with_local_tls_listener() {
    let certificate = rcgen::generate_simple_self_signed(vec!["localhost".to_owned()]).unwrap();
    let expected = certificate.cert.der().to_vec();
    let key = rustls::pki_types::PrivatePkcs8KeyDer::from(certificate.signing_key.serialize_der());
    let config = rustls::ServerConfig::builder_with_provider(Arc::new(
        rustls::crypto::ring::default_provider(),
    ))
    .with_safe_default_protocol_versions()
    .unwrap()
    .with_no_client_auth()
    .with_single_cert(vec![certificate.cert.der().clone()], key.into())
    .unwrap();
    let acceptor = tokio_rustls::TlsAcceptor::from(Arc::new(config));
    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let address = listener.local_addr().unwrap();
    let server = tokio::spawn(async move {
        let (stream, _) = listener.accept().await.unwrap();
        acceptor.accept(stream).await.unwrap();
    });

    let observed = SystemCertificateProbe::new()
        .leaf_certificate(address.ip(), address.port(), "localhost")
        .await
        .unwrap();
    assert_eq!(observed, expected);
    server.await.unwrap();
}

#[tokio::test]
async fn system_website_probe_collects_local_response_without_redirecting() {
    use tokio::io::{AsyncReadExt, AsyncWriteExt};

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let address = listener.local_addr().unwrap();
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.unwrap();
        let mut request = [0_u8; 1024];
        let _ = stream.read(&mut request).await.unwrap();
        stream
            .write_all(b"HTTP/1.1 302 Found\r\nLocation: http://127.0.0.1:1/\r\nServer: local-test\r\nContent-Type: text/html\r\nContent-Length: 40\r\nConnection: close\r\n\r\n<html><title> Local Test </title></html>")
            .await
            .unwrap();
    });
    let snapshot = SystemWebsiteProbe
        .fetch(address.ip(), address.port(), "127.0.0.1", "/")
        .await
        .unwrap();
    assert_eq!(snapshot.status_code, 302);
    assert_eq!(snapshot.title, "Local Test");
    assert_eq!(snapshot.server, "local-test");
    server.await.unwrap();
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
