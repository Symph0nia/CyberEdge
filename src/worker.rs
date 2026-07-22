use std::{
    collections::BTreeSet,
    net::IpAddr,
    sync::Arc,
    time::{SystemTime, UNIX_EPOCH},
};

use async_trait::async_trait;
use serde::Deserialize;
use serde_json::json;
use sha2::{Digest, Sha256};
use tokio::net::lookup_host;
use uuid::Uuid;

use crate::{
    proto::{Asset, Evidence, Observation, ScopeTarget, TargetKind},
    repository::{DiscoveryRecord, Repository, RepositoryError},
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
pub trait CertificateSource: Send + Sync {
    async fn discover(&self, domain: &str) -> Result<Vec<String>, String>;
}

pub struct CrtShSource {
    client: reqwest::Client,
}

impl CrtShSource {
    pub fn new() -> Result<Self, reqwest::Error> {
        Ok(Self {
            client: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(15))
                .user_agent("CyberEdge/0.1 passive-inventory")
                .build()?,
        })
    }
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
}

impl DiscoveryWorker {
    pub fn new(repository: Arc<dyn Repository>, resolver: Arc<dyn DnsResolver>) -> Self {
        Self {
            repository,
            resolver,
            certificate_source: None,
        }
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
            "policy_passive_dns" | "policy_passive_inventory"
        ) {
            self.repository.fail_task(&claimed.task.id, now()).await?;
            return Ok(true);
        }

        let mut records = Vec::new();
        for target in &claimed.scope.targets {
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
            _ => Vec::new(),
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
    }
}

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
