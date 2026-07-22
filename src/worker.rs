use std::{
    collections::BTreeSet,
    net::{IpAddr, SocketAddr},
    sync::Arc,
    time::{SystemTime, UNIX_EPOCH},
};

use async_trait::async_trait;
use rustls::{
    DigitallySignedStruct, SignatureScheme,
    client::danger::{HandshakeSignatureValid, ServerCertVerified, ServerCertVerifier},
    crypto::{CryptoProvider, verify_tls12_signature, verify_tls13_signature},
    pki_types::{CertificateDer, ServerName, UnixTime},
};
use serde::Deserialize;
use serde_json::json;
use sha2::{Digest, Sha256};
use tokio::net::lookup_host;
use tokio::{net::TcpStream, time::timeout};
use tokio_rustls::TlsConnector;
use uuid::Uuid;
use x509_parser::{extensions::GeneralName, parse_x509_certificate};

use crate::{
    proto::{
        Asset, Certificate, Evidence, Finding, FindingSeverity, FindingState, Observation,
        ScopeTarget, Service, TargetKind, Website,
    },
    repository::{DiscoveryRecord, FindingEvaluation, Repository, RepositoryError},
};

#[async_trait]
pub trait DnsResolver: Send + Sync {
    async fn resolve(&self, domain: &str) -> Result<Vec<IpAddr>, String>;
}

pub struct SystemDnsResolver;

#[async_trait]
impl DnsResolver for SystemDnsResolver {
    async fn resolve(&self, domain: &str) -> Result<Vec<IpAddr>, String> {
        let addresses = lookup_host((domain, 0))
            .await
            .map_err(|error| error.to_string())?
            .map(|socket| socket.ip())
            .collect::<BTreeSet<_>>()
            .into_iter()
            .collect();
        Ok(addresses)
    }
}

#[async_trait]
pub trait PortConnector: Send + Sync {
    async fn connect(&self, address: IpAddr, port: u16) -> Result<bool, String>;
}

pub struct SystemPortConnector;

#[async_trait]
impl PortConnector for SystemPortConnector {
    async fn connect(&self, address: IpAddr, port: u16) -> Result<bool, String> {
        match timeout(
            std::time::Duration::from_millis(750),
            TcpStream::connect((address, port)),
        )
        .await
        {
            Ok(Ok(_)) => Ok(true),
            Ok(Err(error))
                if matches!(
                    error.kind(),
                    std::io::ErrorKind::ConnectionRefused
                        | std::io::ErrorKind::TimedOut
                        | std::io::ErrorKind::HostUnreachable
                        | std::io::ErrorKind::NetworkUnreachable
                ) =>
            {
                Ok(false)
            }
            Ok(Err(error)) => Err(error.to_string()),
            Err(_) => Ok(false),
        }
    }
}

pub struct WebSnapshot {
    pub url: String,
    pub status_code: u16,
    pub title: String,
    pub server: String,
    pub content_type: String,
    pub body: Vec<u8>,
}

#[async_trait]
pub trait WebsiteProbe: Send + Sync {
    async fn fetch(
        &self,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<WebSnapshot, String>;
}

pub struct SystemWebsiteProbe;

#[async_trait]
impl WebsiteProbe for SystemWebsiteProbe {
    async fn fetch(
        &self,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<WebSnapshot, String> {
        install_crypto_provider();
        let scheme = if matches!(port, 443 | 8443) {
            "https"
        } else {
            "http"
        };
        let host = if server_name.contains(':') {
            format!("[{server_name}]")
        } else {
            server_name.to_owned()
        };
        let url = format!("{scheme}://{host}:{port}/");
        let mut builder = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(5))
            .redirect(reqwest::redirect::Policy::none())
            .danger_accept_invalid_certs(true)
            .user_agent("CyberEdge/0.1 authorized-observer");
        if server_name.parse::<IpAddr>().is_err() {
            builder = builder.resolve(server_name, SocketAddr::new(address, port));
        }
        let mut response = builder
            .build()
            .map_err(|error| error.to_string())?
            .get(&url)
            .send()
            .await
            .map_err(|error| error.to_string())?;
        let status_code = response.status().as_u16();
        let server = response
            .headers()
            .get(reqwest::header::SERVER)
            .and_then(|value| value.to_str().ok())
            .unwrap_or_default()
            .chars()
            .take(256)
            .collect::<String>();
        let content_type = response
            .headers()
            .get(reqwest::header::CONTENT_TYPE)
            .and_then(|value| value.to_str().ok())
            .unwrap_or("application/octet-stream")
            .chars()
            .take(256)
            .collect::<String>();
        let mut body = Vec::new();
        while let Some(chunk) = response.chunk().await.map_err(|error| error.to_string())? {
            if body.len() + chunk.len() > MAX_WEB_BODY_BYTES {
                return Err(format!("response body exceeds {MAX_WEB_BODY_BYTES} bytes"));
            }
            body.extend_from_slice(&chunk);
        }
        Ok(WebSnapshot {
            url,
            status_code,
            title: html_title(&body),
            server,
            content_type,
            body,
        })
    }
}

#[async_trait]
pub trait CertificateProbe: Send + Sync {
    async fn leaf_certificate(
        &self,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<Vec<u8>, String>;
}

pub struct SystemCertificateProbe {
    connector: TlsConnector,
}

impl SystemCertificateProbe {
    pub fn new() -> Self {
        let provider = Arc::new(rustls::crypto::ring::default_provider());
        let verifier = Arc::new(CollectCertificateVerifier {
            provider: provider.clone(),
        });
        let config = rustls::ClientConfig::builder_with_provider(provider)
            .with_safe_default_protocol_versions()
            .expect("ring provider supports default TLS versions")
            .dangerous()
            .with_custom_certificate_verifier(verifier)
            .with_no_client_auth();
        Self {
            connector: TlsConnector::from(Arc::new(config)),
        }
    }
}

impl Default for SystemCertificateProbe {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug)]
struct CollectCertificateVerifier {
    provider: Arc<CryptoProvider>,
}

impl ServerCertVerifier for CollectCertificateVerifier {
    fn verify_server_cert(
        &self,
        _end_entity: &CertificateDer<'_>,
        _intermediates: &[CertificateDer<'_>],
        _server_name: &ServerName<'_>,
        _ocsp_response: &[u8],
        _now: UnixTime,
    ) -> Result<ServerCertVerified, rustls::Error> {
        Ok(ServerCertVerified::assertion())
    }

    fn verify_tls12_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, rustls::Error> {
        verify_tls12_signature(
            message,
            cert,
            dss,
            &self.provider.signature_verification_algorithms,
        )
    }

    fn verify_tls13_signature(
        &self,
        message: &[u8],
        cert: &CertificateDer<'_>,
        dss: &DigitallySignedStruct,
    ) -> Result<HandshakeSignatureValid, rustls::Error> {
        verify_tls13_signature(
            message,
            cert,
            dss,
            &self.provider.signature_verification_algorithms,
        )
    }

    fn supported_verify_schemes(&self) -> Vec<SignatureScheme> {
        self.provider
            .signature_verification_algorithms
            .supported_schemes()
    }
}

#[async_trait]
impl CertificateProbe for SystemCertificateProbe {
    async fn leaf_certificate(
        &self,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> Result<Vec<u8>, String> {
        let stream = timeout(
            std::time::Duration::from_secs(2),
            TcpStream::connect((address, port)),
        )
        .await
        .map_err(|_| "TLS connect timed out".to_owned())?
        .map_err(|error| error.to_string())?;
        let server_name =
            ServerName::try_from(server_name.to_owned()).map_err(|error| error.to_string())?;
        let stream = timeout(
            std::time::Duration::from_secs(3),
            self.connector.connect(server_name, stream),
        )
        .await
        .map_err(|_| "TLS handshake timed out".to_owned())?
        .map_err(|error| error.to_string())?;
        stream
            .get_ref()
            .1
            .peer_certificates()
            .and_then(|certificates| certificates.first())
            .map(|certificate| certificate.as_ref().to_vec())
            .ok_or_else(|| "TLS peer did not provide a certificate".to_owned())
    }
}

#[async_trait]
pub trait CertificateSource: Send + Sync {
    async fn discover(&self, domain: &str) -> Result<Vec<String>, String>;
}

pub struct CrtShSource {
    client: reqwest::Client,
}

impl CrtShSource {
    pub fn new() -> Result<Self, reqwest::Error> {
        install_crypto_provider();
        Ok(Self {
            client: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(15))
                .user_agent("CyberEdge/0.1 passive-inventory")
                .build()?,
        })
    }
}

fn install_crypto_provider() {
    let _ = rustls::crypto::ring::default_provider().install_default();
}

#[derive(Deserialize)]
struct CrtShEntry {
    name_value: String,
}

#[async_trait]
impl CertificateSource for CrtShSource {
    async fn discover(&self, domain: &str) -> Result<Vec<String>, String> {
        let response = self
            .client
            .get("https://crt.sh/")
            .query(&[("q", format!("%.{domain}")), ("output", "json".to_owned())])
            .send()
            .await
            .map_err(|error| error.to_string())?
            .error_for_status()
            .map_err(|error| error.to_string())?;
        let entries = response
            .json::<Vec<CrtShEntry>>()
            .await
            .map_err(|error| error.to_string())?;
        Ok(entries
            .into_iter()
            .flat_map(|entry| {
                entry
                    .name_value
                    .lines()
                    .map(str::to_owned)
                    .collect::<Vec<_>>()
            })
            .collect())
    }
}

pub struct DiscoveryWorker {
    repository: Arc<dyn Repository>,
    resolver: Arc<dyn DnsResolver>,
    certificate_source: Option<Arc<dyn CertificateSource>>,
    port_connector: Option<Arc<dyn PortConnector>>,
    certificate_probe: Option<Arc<dyn CertificateProbe>>,
    website_probe: Option<Arc<dyn WebsiteProbe>>,
    service_ports: Vec<u16>,
}

impl DiscoveryWorker {
    pub fn new(repository: Arc<dyn Repository>, resolver: Arc<dyn DnsResolver>) -> Self {
        Self {
            repository,
            resolver,
            certificate_source: None,
            port_connector: None,
            certificate_probe: None,
            website_probe: None,
            service_ports: Vec::new(),
        }
    }

    pub fn with_certificate_probe(mut self, probe: Arc<dyn CertificateProbe>) -> Self {
        self.certificate_probe = Some(probe);
        self
    }

    pub fn with_website_probe(mut self, probe: Arc<dyn WebsiteProbe>) -> Self {
        self.website_probe = Some(probe);
        self
    }

    pub fn with_port_connector(
        mut self,
        connector: Arc<dyn PortConnector>,
        ports: Vec<u16>,
    ) -> Self {
        self.port_connector = Some(connector);
        self.service_ports = ports;
        self
    }

    pub fn with_certificate_source(mut self, source: Arc<dyn CertificateSource>) -> Self {
        self.certificate_source = Some(source);
        self
    }

    pub async fn run_once(&self) -> Result<bool, RepositoryError> {
        let Some(claimed) = self.repository.claim_task(now()).await? else {
            return Ok(false);
        };
        if !matches!(
            claimed.task.policy_id.as_str(),
            "policy_passive_dns" | "policy_passive_inventory" | "policy_service_baseline"
        ) {
            self.repository.fail_task(&claimed.task.id, now()).await?;
            return Ok(true);
        }

        let mut records = Vec::new();
        for target in &claimed.scope.targets {
            if claimed.task.policy_id == "policy_service_baseline" {
                records.extend(
                    self.discover_active_target(&claimed.task.id, &claimed.scope.id, target)
                        .await,
                );
                continue;
            }
            records.extend(
                self.discover_target(&claimed.task.id, &claimed.scope.id, target)
                    .await,
            );
            if claimed.task.policy_id == "policy_passive_inventory" {
                records.extend(
                    self.discover_certificates(&claimed.task.id, &claimed.scope.id, target)
                        .await,
                );
            }
        }
        self.repository
            .complete_task(&claimed.task.id, records, now())
            .await?;
        Ok(true)
    }

    async fn discover_active_target(
        &self,
        task_id: &str,
        scope_id: &str,
        target: &ScopeTarget,
    ) -> Vec<DiscoveryRecord> {
        match TargetKind::try_from(target.kind).unwrap_or(TargetKind::Unspecified) {
            TargetKind::Ip => match target.value.parse::<IpAddr>() {
                Ok(address) => {
                    let mut records = vec![record(
                        task_id,
                        scope_id,
                        TargetKind::Ip,
                        &target.value,
                        "scope.seed",
                        json!({"source": "authorized_scope", "value": target.value}),
                    )];
                    records.extend(
                        self.discover_services(task_id, scope_id, address, &target.value)
                            .await,
                    );
                    records
                }
                Err(error) => vec![record(
                    task_id,
                    scope_id,
                    TargetKind::Ip,
                    &target.value,
                    "tcp.error",
                    json!({"address": target.value, "error": error.to_string()}),
                )],
            },
            TargetKind::Domain => match self.resolver.resolve(&target.value).await {
                Ok(addresses) => {
                    let values = addresses
                        .iter()
                        .map(ToString::to_string)
                        .collect::<Vec<_>>();
                    let mut records = vec![record(
                        task_id,
                        scope_id,
                        TargetKind::Domain,
                        &target.value,
                        "dns.addresses",
                        json!({"domain": target.value, "addresses": values}),
                    )];
                    for address in addresses {
                        records.extend(
                            self.discover_services(task_id, scope_id, address, &target.value)
                                .await,
                        );
                    }
                    records
                }
                Err(error) => vec![record(
                    task_id,
                    scope_id,
                    TargetKind::Domain,
                    &target.value,
                    "dns.error",
                    json!({"domain": target.value, "error": error}),
                )],
            },
            kind => vec![unsupported_target_record(
                task_id,
                scope_id,
                kind,
                &target.value,
            )],
        }
    }

    async fn discover_services(
        &self,
        task_id: &str,
        scope_id: &str,
        address: IpAddr,
        server_name: &str,
    ) -> Vec<DiscoveryRecord> {
        let Some(connector) = &self.port_connector else {
            return vec![record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "tcp.error",
                json!({"address": address, "error": "port connector unavailable"}),
            )];
        };
        let mut probes = tokio::task::JoinSet::new();
        for port in self.service_ports.iter().copied() {
            let connector = connector.clone();
            probes.spawn(async move { (port, connector.connect(address, port).await) });
        }
        let mut results = Vec::with_capacity(self.service_ports.len());
        while let Some(result) = probes.join_next().await {
            match result {
                Ok(result) => results.push(result),
                Err(error) => results.push((0, Err(error.to_string()))),
            }
        }
        results.sort_by_key(|(port, _)| *port);

        let mut records = Vec::new();
        for (port, result) in results {
            match result {
                Ok(true) => {
                    records.push(service_record(task_id, scope_id, address, port));
                    if matches!(port, 443 | 8443) {
                        records.push(
                            self.discover_certificate(
                                task_id,
                                scope_id,
                                address,
                                port,
                                server_name,
                            )
                            .await,
                        );
                    }
                    if matches!(port, 80 | 443 | 8080 | 8443) {
                        records.push(
                            self.discover_website(task_id, scope_id, address, port, server_name)
                                .await,
                        );
                    }
                }
                Ok(false) => {}
                Err(error) => records.push(record(
                    task_id,
                    scope_id,
                    TargetKind::Ip,
                    &address.to_string(),
                    "tcp.error",
                    json!({"address": address, "port": port, "error": error}),
                )),
            }
        }
        if records.is_empty() {
            records.push(record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "tcp.no_open_ports",
                json!({"address": address, "ports": self.service_ports}),
            ));
        }
        records
    }

    async fn discover_website(
        &self,
        task_id: &str,
        scope_id: &str,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> DiscoveryRecord {
        let Some(probe) = &self.website_probe else {
            return record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "http.error",
                json!({"address": address, "port": port, "error": "website probe unavailable"}),
            );
        };
        match probe.fetch(address, port, server_name).await {
            Ok(snapshot) => website_record(task_id, scope_id, address, port, snapshot),
            Err(error) => record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "http.error",
                json!({"address": address, "port": port, "error": error}),
            ),
        }
    }

    async fn discover_certificate(
        &self,
        task_id: &str,
        scope_id: &str,
        address: IpAddr,
        port: u16,
        server_name: &str,
    ) -> DiscoveryRecord {
        let Some(probe) = &self.certificate_probe else {
            return record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "tls.error",
                json!({"address": address, "port": port, "error": "certificate probe unavailable"}),
            );
        };
        match probe.leaf_certificate(address, port, server_name).await {
            Ok(der) => {
                certificate_record(task_id, scope_id, address, port, der).unwrap_or_else(|error| {
                    record(
                        task_id,
                        scope_id,
                        TargetKind::Ip,
                        &address.to_string(),
                        "tls.error",
                        json!({"address": address, "port": port, "error": error}),
                    )
                })
            }
            Err(error) => record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &address.to_string(),
                "tls.error",
                json!({"address": address, "port": port, "error": error}),
            ),
        }
    }

    async fn discover_certificates(
        &self,
        task_id: &str,
        scope_id: &str,
        target: &ScopeTarget,
    ) -> Vec<DiscoveryRecord> {
        if TargetKind::try_from(target.kind) != Ok(TargetKind::Domain) {
            return Vec::new();
        }
        let Some(source) = &self.certificate_source else {
            return vec![record(
                task_id,
                scope_id,
                TargetKind::Domain,
                &target.value,
                "ct.error",
                json!({"domain": target.value, "error": "certificate source unavailable"}),
            )];
        };
        match source.discover(&target.value).await {
            Ok(names) => normalize_certificate_names(&target.value, names)
                .into_iter()
                .map(|domain| {
                    record(
                        task_id,
                        scope_id,
                        TargetKind::Domain,
                        &domain,
                        "ct.discovered_domain",
                        json!({"source": "crt.sh", "root_domain": target.value, "domain": domain}),
                    )
                })
                .collect(),
            Err(error) => vec![record(
                task_id,
                scope_id,
                TargetKind::Domain,
                &target.value,
                "ct.error",
                json!({"domain": target.value, "error": error}),
            )],
        }
    }

    async fn discover_target(
        &self,
        task_id: &str,
        scope_id: &str,
        target: &ScopeTarget,
    ) -> Vec<DiscoveryRecord> {
        match TargetKind::try_from(target.kind).unwrap_or(TargetKind::Unspecified) {
            TargetKind::Domain => self.discover_domain(task_id, scope_id, &target.value).await,
            TargetKind::Ip => vec![record(
                task_id,
                scope_id,
                TargetKind::Ip,
                &target.value,
                "scope.seed",
                json!({"source": "authorized_scope", "value": target.value}),
            )],
            kind => vec![unsupported_target_record(
                task_id,
                scope_id,
                kind,
                &target.value,
            )],
        }
    }

    async fn discover_domain(
        &self,
        task_id: &str,
        scope_id: &str,
        domain: &str,
    ) -> Vec<DiscoveryRecord> {
        match self.resolver.resolve(domain).await {
            Ok(addresses) if addresses.is_empty() => vec![record(
                task_id,
                scope_id,
                TargetKind::Domain,
                domain,
                "dns.no_data",
                json!({"domain": domain, "addresses": []}),
            )],
            Ok(addresses) => {
                let values = addresses
                    .iter()
                    .map(ToString::to_string)
                    .collect::<Vec<_>>();
                let mut records = vec![record(
                    task_id,
                    scope_id,
                    TargetKind::Domain,
                    domain,
                    "dns.addresses",
                    json!({"domain": domain, "addresses": values}),
                )];
                records.extend(addresses.into_iter().map(|address| {
                    record(
                        task_id,
                        scope_id,
                        TargetKind::Ip,
                        &address.to_string(),
                        "dns.discovered_ip",
                        json!({"domain": domain, "address": address}),
                    )
                }));
                records
            }
            Err(error) => vec![record(
                task_id,
                scope_id,
                TargetKind::Domain,
                domain,
                "dns.error",
                json!({"domain": domain, "error": error}),
            )],
        }
    }
}

const MAX_CERTIFICATE_NAMES: usize = 1_000;

fn normalize_certificate_names(root: &str, names: Vec<String>) -> Vec<String> {
    let root = root.trim().trim_end_matches('.').to_ascii_lowercase();
    names
        .into_iter()
        .filter_map(|name| {
            let name = name
                .trim()
                .trim_start_matches("*.")
                .trim_end_matches('.')
                .to_ascii_lowercase();
            ((!name.is_empty())
                && (name == root || name.ends_with(&format!(".{root}")))
                && name.split('.').all(|label| {
                    !label.is_empty()
                        && label.len() <= 63
                        && label
                            .bytes()
                            .all(|byte| byte.is_ascii_alphanumeric() || byte == b'-')
                }))
            .then_some(name)
        })
        .collect::<BTreeSet<_>>()
        .into_iter()
        .take(MAX_CERTIFICATE_NAMES)
        .collect()
}

fn record(
    task_id: &str,
    scope_id: &str,
    kind: TargetKind,
    value: &str,
    observation_type: &str,
    payload: serde_json::Value,
) -> DiscoveryRecord {
    let timestamp = now();
    let content = serde_json::to_vec(&payload).expect("JSON evidence is serializable");
    let evidence_hash = hex_hash(&content);
    let asset_id = stable_id(
        "asset",
        format!("{scope_id}:{}:{value}", i32::from(kind)).as_bytes(),
    );
    let evidence_id = format!("evidence_{evidence_hash}");
    DiscoveryRecord {
        asset: Asset {
            id: asset_id.clone(),
            scope_id: scope_id.to_owned(),
            kind: kind.into(),
            value: value.to_owned(),
            first_seen_at: Some(timestamp),
            last_seen_at: Some(timestamp),
        },
        service: None,
        certificate: None,
        website: None,
        observation: Observation {
            id: format!("observation_{}", Uuid::now_v7()),
            task_id: task_id.to_owned(),
            asset_id,
            observation_type: observation_type.to_owned(),
            value_json: String::from_utf8(content.clone()).expect("JSON is UTF-8"),
            evidence_id: evidence_id.clone(),
            observed_at: Some(timestamp),
        },
        evidence: Evidence {
            id: evidence_id,
            media_type: "application/json".to_owned(),
            sha256: evidence_hash,
            content,
            created_at: Some(timestamp),
        },
        finding_evaluations: Vec::new(),
        findings: Vec::new(),
    }
}

fn service_record(task_id: &str, scope_id: &str, address: IpAddr, port: u16) -> DiscoveryRecord {
    let value = address.to_string();
    let mut record = record(
        task_id,
        scope_id,
        TargetKind::Ip,
        &value,
        "tcp.open",
        json!({"address": address, "transport": "tcp", "port": port}),
    );
    let timestamp = record
        .observation
        .observed_at
        .expect("observation timestamp exists");
    record.service = Some(Service {
        id: stable_id(
            "service",
            format!("{}:tcp:{port}", record.asset.id).as_bytes(),
        ),
        asset_id: record.asset.id.clone(),
        transport: "tcp".to_owned(),
        port: u32::from(port),
        service_hint: service_hint(port).to_owned(),
        first_seen_at: Some(timestamp),
        last_seen_at: Some(timestamp),
    });
    record
}

fn certificate_record(
    task_id: &str,
    scope_id: &str,
    address: IpAddr,
    port: u16,
    der: Vec<u8>,
) -> Result<DiscoveryRecord, String> {
    let (_, parsed) = parse_x509_certificate(&der).map_err(|error| error.to_string())?;
    let dns_names = parsed
        .subject_alternative_name()
        .map_err(|error| error.to_string())?
        .into_iter()
        .flat_map(|extension| extension.value.general_names.iter())
        .filter_map(|name| match name {
            GeneralName::DNSName(value) => Some((*value).to_owned()),
            _ => None,
        })
        .collect::<BTreeSet<_>>()
        .into_iter()
        .collect::<Vec<_>>();
    let timestamp = now();
    let value = address.to_string();
    let asset_id = stable_id(
        "asset",
        format!("{scope_id}:{}:{value}", i32::from(TargetKind::Ip)).as_bytes(),
    );
    let service_id = stable_id("service", format!("{asset_id}:tcp:{port}").as_bytes());
    let sha256 = hex_hash(&der);
    let observation_id = format!("observation_{}", Uuid::now_v7());
    let evidence_id = format!("evidence_{sha256}");
    let certificate = Certificate {
        id: stable_id("certificate", format!("{service_id}:{sha256}").as_bytes()),
        service_id: service_id.clone(),
        sha256: sha256.clone(),
        subject: parsed.subject().to_string(),
        issuer: parsed.issuer().to_string(),
        dns_names: dns_names.clone(),
        not_before: Some(prost_types::Timestamp {
            seconds: parsed.validity().not_before.timestamp(),
            nanos: 0,
        }),
        not_after: Some(prost_types::Timestamp {
            seconds: parsed.validity().not_after.timestamp(),
            nanos: 0,
        }),
        first_seen_at: Some(timestamp),
        last_seen_at: Some(timestamp),
    };
    let (finding_evaluations, findings) = certificate_validity_detections(
        task_id,
        scope_id,
        &asset_id,
        &service_id,
        &observation_id,
        &evidence_id,
        &certificate,
        timestamp,
    );
    let payload = json!({
        "address": address,
        "port": port,
        "sha256": sha256,
        "subject": certificate.subject,
        "issuer": certificate.issuer,
        "dns_names": dns_names,
        "not_before_unix": certificate.not_before.as_ref().map(|value| value.seconds),
        "not_after_unix": certificate.not_after.as_ref().map(|value| value.seconds),
    });
    Ok(DiscoveryRecord {
        asset: Asset {
            id: asset_id.clone(),
            scope_id: scope_id.to_owned(),
            kind: TargetKind::Ip.into(),
            value,
            first_seen_at: Some(timestamp),
            last_seen_at: Some(timestamp),
        },
        service: Some(Service {
            id: service_id,
            asset_id: asset_id.clone(),
            transport: "tcp".to_owned(),
            port: u32::from(port),
            service_hint: "https".to_owned(),
            first_seen_at: Some(timestamp),
            last_seen_at: Some(timestamp),
        }),
        certificate: Some(certificate),
        website: None,
        observation: Observation {
            id: observation_id,
            task_id: task_id.to_owned(),
            asset_id,
            observation_type: "tls.certificate".to_owned(),
            value_json: payload.to_string(),
            evidence_id: evidence_id.clone(),
            observed_at: Some(timestamp),
        },
        evidence: Evidence {
            id: evidence_id,
            media_type: "application/pkix-cert".to_owned(),
            sha256,
            content: der,
            created_at: Some(timestamp),
        },
        finding_evaluations,
        findings,
    })
}

#[allow(clippy::too_many_arguments)]
fn certificate_validity_detections(
    task_id: &str,
    scope_id: &str,
    asset_id: &str,
    service_id: &str,
    observation_id: &str,
    evidence_id: &str,
    certificate: &Certificate,
    timestamp: prost_types::Timestamp,
) -> (Vec<FindingEvaluation>, Vec<Finding>) {
    const EXPIRING_WINDOW_SECONDS: i64 = 30 * 24 * 60 * 60;
    const DETECTOR: &str = "cyberedge-tls";
    const EXPIRED_RULE: &str = "tls-certificate-expired-v1";
    const EXPIRING_RULE: &str = "tls-certificate-expiring-v1";

    let fingerprint = |rule_id: &'static str| {
        let fingerprint = hex_hash(format!("{rule_id}:{service_id}").as_bytes());
        FindingEvaluation {
            asset_id: asset_id.to_owned(),
            detector: DETECTOR,
            rule_id,
            fingerprint,
        }
    };
    let expired = fingerprint(EXPIRED_RULE);
    let expiring = fingerprint(EXPIRING_RULE);
    let evaluations = vec![expired, expiring];
    let Some(not_after) = certificate.not_after else {
        return (evaluations, Vec::new());
    };

    let selected = if not_after.seconds <= timestamp.seconds {
        Some((
            &evaluations[0],
            "TLS certificate expired",
            FindingSeverity::High,
        ))
    } else if not_after.seconds <= timestamp.seconds + EXPIRING_WINDOW_SECONDS {
        Some((
            &evaluations[1],
            "TLS certificate expires within 30 days",
            FindingSeverity::Medium,
        ))
    } else {
        None
    };
    let Some((evaluation, title, severity)) = selected else {
        return (evaluations, Vec::new());
    };
    let finding = Finding {
        id: stable_id(
            "finding",
            format!(
                "{scope_id}:{}:{}:{asset_id}:{}",
                evaluation.detector, evaluation.rule_id, evaluation.fingerprint
            )
            .as_bytes(),
        ),
        scope_id: scope_id.to_owned(),
        task_id: task_id.to_owned(),
        asset_id: asset_id.to_owned(),
        observation_id: observation_id.to_owned(),
        evidence_id: evidence_id.to_owned(),
        detector: evaluation.detector.to_owned(),
        rule_id: evaluation.rule_id.to_owned(),
        title: title.to_owned(),
        description: format!(
            "The retained DER certificate for service {service_id} expires at Unix timestamp {}.",
            not_after.seconds
        ),
        severity: severity.into(),
        state: FindingState::Open.into(),
        fingerprint: evaluation.fingerprint.clone(),
        first_seen_at: Some(timestamp),
        last_seen_at: Some(timestamp),
    };
    (evaluations, vec![finding])
}

fn website_record(
    task_id: &str,
    scope_id: &str,
    address: IpAddr,
    port: u16,
    snapshot: WebSnapshot,
) -> DiscoveryRecord {
    let timestamp = now();
    let value = address.to_string();
    let asset_id = stable_id(
        "asset",
        format!("{scope_id}:{}:{value}", i32::from(TargetKind::Ip)).as_bytes(),
    );
    let service_id = stable_id("service", format!("{asset_id}:tcp:{port}").as_bytes());
    let content_sha256 = hex_hash(&snapshot.body);
    let evidence_id = format!("evidence_{content_sha256}");
    let observation_id = format!("observation_{}", Uuid::now_v7());
    let website = Website {
        id: stable_id("website", service_id.as_bytes()),
        service_id: service_id.clone(),
        url: snapshot.url.clone(),
        status_code: u32::from(snapshot.status_code),
        title: snapshot.title.clone(),
        server: snapshot.server.clone(),
        content_type: snapshot.content_type.clone(),
        content_sha256: content_sha256.clone(),
        first_seen_at: Some(timestamp),
        last_seen_at: Some(timestamp),
    };
    let evidence_media_type = website.content_type.clone();
    let (finding_evaluation, finding) = directory_listing_detection(
        task_id,
        scope_id,
        &asset_id,
        &observation_id,
        &evidence_id,
        &snapshot,
        timestamp,
    );
    let payload = json!({
        "address": address, "port": port, "url": snapshot.url,
        "status_code": snapshot.status_code, "title": snapshot.title,
        "server": snapshot.server, "content_type": snapshot.content_type,
        "content_sha256": content_sha256,
    });
    DiscoveryRecord {
        asset: Asset {
            id: asset_id.clone(),
            scope_id: scope_id.to_owned(),
            kind: TargetKind::Ip.into(),
            value,
            first_seen_at: Some(timestamp),
            last_seen_at: Some(timestamp),
        },
        service: Some(Service {
            id: service_id,
            asset_id: asset_id.clone(),
            transport: "tcp".to_owned(),
            port: u32::from(port),
            service_hint: service_hint(port).to_owned(),
            first_seen_at: Some(timestamp),
            last_seen_at: Some(timestamp),
        }),
        certificate: None,
        website: Some(website),
        observation: Observation {
            id: observation_id,
            task_id: task_id.to_owned(),
            asset_id,
            observation_type: "http.response".to_owned(),
            value_json: payload.to_string(),
            evidence_id: evidence_id.clone(),
            observed_at: Some(timestamp),
        },
        evidence: Evidence {
            id: evidence_id,
            media_type: evidence_media_type,
            sha256: content_sha256,
            content: snapshot.body,
            created_at: Some(timestamp),
        },
        finding_evaluations: vec![finding_evaluation],
        findings: finding.into_iter().collect(),
    }
}

fn directory_listing_detection(
    task_id: &str,
    scope_id: &str,
    asset_id: &str,
    observation_id: &str,
    evidence_id: &str,
    snapshot: &WebSnapshot,
    timestamp: prost_types::Timestamp,
) -> (FindingEvaluation, Option<Finding>) {
    let detector = "cyberedge-http";
    let rule_id = "http-directory-listing-v1";
    let fingerprint = hex_hash(format!("{rule_id}:{}", snapshot.url).as_bytes());
    let evaluation = FindingEvaluation {
        asset_id: asset_id.to_owned(),
        detector,
        rule_id,
        fingerprint: fingerprint.clone(),
    };
    let title = snapshot.title.trim().to_ascii_lowercase();
    let body = String::from_utf8_lossy(&snapshot.body).to_ascii_lowercase();
    let listing_title = title.starts_with("index of /")
        || body.contains("<title>index of /")
        || body.contains("<h1>index of /");
    let listing_marker = body.contains("parent directory") || body.contains("?c=n;o=");
    if snapshot.status_code != 200 || !listing_title || !listing_marker {
        return (evaluation, None);
    }

    let finding = Finding {
        id: stable_id(
            "finding",
            format!("{scope_id}:{detector}:{rule_id}:{asset_id}:{fingerprint}").as_bytes(),
        ),
        scope_id: scope_id.to_owned(),
        task_id: task_id.to_owned(),
        asset_id: asset_id.to_owned(),
        observation_id: observation_id.to_owned(),
        evidence_id: evidence_id.to_owned(),
        detector: detector.to_owned(),
        rule_id: rule_id.to_owned(),
        title: "HTTP directory listing exposed".to_owned(),
        description: format!(
            "The HTTP response at {} exposes a directory index backed by the retained response body.",
            snapshot.url
        ),
        severity: FindingSeverity::Medium.into(),
        state: FindingState::Open.into(),
        fingerprint,
        first_seen_at: Some(timestamp),
        last_seen_at: Some(timestamp),
    };
    (evaluation, Some(finding))
}

const MAX_WEB_BODY_BYTES: usize = 1_048_576;

fn html_title(body: &[u8]) -> String {
    let text = String::from_utf8_lossy(body);
    let lower = text.to_ascii_lowercase();
    let Some(start) = lower.find("<title") else {
        return String::new();
    };
    let Some(open) = lower[start..].find('>').map(|offset| start + offset + 1) else {
        return String::new();
    };
    let Some(close) = lower[open..].find("</title>").map(|offset| open + offset) else {
        return String::new();
    };
    text[open..close]
        .split_whitespace()
        .collect::<Vec<_>>()
        .join(" ")
        .chars()
        .take(512)
        .collect()
}

fn unsupported_target_record(
    task_id: &str,
    scope_id: &str,
    kind: TargetKind,
    value: &str,
) -> DiscoveryRecord {
    record(
        task_id,
        scope_id,
        kind,
        value,
        "policy.error",
        json!({"target": value, "kind": i32::from(kind), "error": "target kind unsupported by policy"}),
    )
}

fn service_hint(port: u16) -> &'static str {
    match port {
        22 => "ssh",
        25 => "smtp",
        53 => "dns",
        80 | 8080 => "http",
        110 => "pop3",
        143 => "imap",
        443 | 8443 => "https",
        445 => "smb",
        3306 => "mysql",
        5432 => "postgresql",
        6379 => "redis",
        _ => "unknown",
    }
}

pub const BASELINE_SERVICE_PORTS: &[u16] = &[
    22, 25, 53, 80, 110, 143, 443, 445, 3306, 5432, 6379, 8080, 8443,
];

fn stable_id(prefix: &str, value: &[u8]) -> String {
    format!("{prefix}_{}", &hex_hash(value)[..32])
}

fn hex_hash(value: &[u8]) -> String {
    Sha256::digest(value)
        .iter()
        .map(|byte| format!("{byte:02x}"))
        .collect()
}

fn now() -> prost_types::Timestamp {
    let duration = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("system time must be after Unix epoch");
    prost_types::Timestamp {
        seconds: duration.as_secs() as i64,
        nanos: duration.subsec_nanos() as i32,
    }
}

#[cfg(test)]
mod detector_tests {
    use super::*;

    fn certificate(not_after: i64) -> Certificate {
        Certificate {
            not_after: Some(prost_types::Timestamp {
                seconds: not_after,
                nanos: 0,
            }),
            ..Certificate::default()
        }
    }

    #[test]
    fn classifies_certificate_validity_windows() {
        let timestamp = prost_types::Timestamp {
            seconds: 1_000_000_000,
            nanos: 0,
        };
        let detect = |not_after| {
            certificate_validity_detections(
                "task",
                "scope",
                "asset",
                "service",
                "observation",
                "evidence",
                &certificate(not_after),
                timestamp,
            )
            .1
        };
        assert_eq!(
            detect(timestamp.seconds - 1)[0].rule_id,
            "tls-certificate-expired-v1"
        );
        assert_eq!(
            detect(timestamp.seconds + 29 * 24 * 60 * 60)[0].rule_id,
            "tls-certificate-expiring-v1"
        );
        assert!(detect(timestamp.seconds + 31 * 24 * 60 * 60).is_empty());
    }
}
