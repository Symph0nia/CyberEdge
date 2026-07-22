use std::{
    collections::BTreeSet,
    net::IpAddr,
    sync::Arc,
    time::{SystemTime, UNIX_EPOCH},
};

use async_trait::async_trait;
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

pub struct DiscoveryWorker {
    repository: Arc<dyn Repository>,
    resolver: Arc<dyn DnsResolver>,
}

impl DiscoveryWorker {
    pub fn new(repository: Arc<dyn Repository>, resolver: Arc<dyn DnsResolver>) -> Self {
        Self {
            repository,
            resolver,
        }
    }

    pub async fn run_once(&self) -> Result<bool, RepositoryError> {
        let Some(claimed) = self.repository.claim_task(now()).await? else {
            return Ok(false);
        };
        if claimed.task.policy_id != "policy_passive_dns" {
            self.repository.fail_task(&claimed.task.id, now()).await?;
            return Ok(true);
        }

        let mut records = Vec::new();
        for target in &claimed.scope.targets {
            records.extend(
                self.discover_target(&claimed.task.id, &claimed.scope.id, target)
                    .await,
            );
        }
        self.repository
            .complete_task(&claimed.task.id, records, now())
            .await?;
        Ok(true)
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
