use std::path::PathBuf;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use tokio::io::{AsyncReadExt, AsyncWriteExt};

const MAX_DOMAINS: usize = 8;
const MAX_DOMAIN_BYTES: usize = 253;
const MAX_REQUEST_BYTES: usize = 4 * 1024;
const MAX_RESPONSE_BYTES: usize = 1024 * 1024;
const MAX_CPE_NAMES: usize = 16;
const MAX_CVE_RESPONSE_BYTES: usize = 10 * 1024 * 1024;
const MAX_REGISTRATION_RESPONSE_BYTES: usize = 2 * 1024 * 1024;

#[async_trait]
pub trait PublicCodeProbe: Send + Sync {
    async fn search(&self, domains: &[String]) -> Result<Vec<u8>, String>;
}

#[async_trait]
pub trait CveProbe: Send + Sync {
    async fn query(&self, cpe_names: &[String]) -> Result<Vec<u8>, String>;
}

#[async_trait]
pub trait RegistrationProbe: Send + Sync {
    async fn lookup(&self, domains: &[String]) -> Result<Vec<u8>, String>;
}

pub struct SocketPublicCodeProbe {
    socket_path: PathBuf,
}

pub struct GitHubPublicCodeProbe {
    client: reqwest::Client,
    token: String,
    api_base: reqwest::Url,
}

pub struct SocketCveProbe {
    socket_path: PathBuf,
}

pub struct NvdCveProbe {
    client: reqwest::Client,
    api_key: Option<String>,
    cve_api: reqwest::Url,
    cpe_api: reqwest::Url,
}

pub struct SocketRegistrationProbe {
    socket_path: PathBuf,
}

pub struct HttpRegistrationProbe {
    client: reqwest::Client,
    endpoint: reqwest::Url,
    token: String,
}

impl SocketPublicCodeProbe {
    pub fn new(socket_path: impl Into<PathBuf>) -> Self {
        Self {
            socket_path: socket_path.into(),
        }
    }
}

impl GitHubPublicCodeProbe {
    pub fn new(token: impl Into<String>) -> Result<Self, String> {
        Self::with_api_base(token, "https://api.github.com/")
    }

    pub fn with_api_base(token: impl Into<String>, api_base: &str) -> Result<Self, String> {
        let token = token.into();
        if token.trim().is_empty() {
            return Err("GitHub token is required".to_owned());
        }
        let api_base = reqwest::Url::parse(api_base).map_err(|error| error.to_string())?;
        let _ = rustls::crypto::ring::default_provider().install_default();
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(15))
            .redirect(reqwest::redirect::Policy::none())
            .user_agent("CyberEdge/0.1 public-code-intelligence")
            .build()
            .map_err(|error| error.to_string())?;
        Ok(Self {
            client,
            token,
            api_base,
        })
    }
}

impl SocketCveProbe {
    pub fn new(socket_path: impl Into<PathBuf>) -> Self {
        Self {
            socket_path: socket_path.into(),
        }
    }
}

impl NvdCveProbe {
    pub fn new(api_key: Option<String>) -> Result<Self, String> {
        Self::with_api_bases(
            api_key,
            "https://services.nvd.nist.gov/rest/json/cves/2.0",
            "https://services.nvd.nist.gov/rest/json/cpes/2.0",
        )
    }

    pub fn with_api_bases(
        api_key: Option<String>,
        cve_api: &str,
        cpe_api: &str,
    ) -> Result<Self, String> {
        let _ = rustls::crypto::ring::default_provider().install_default();
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .redirect(reqwest::redirect::Policy::none())
            .user_agent("CyberEdge/0.1 exact-cpe-intelligence")
            .build()
            .map_err(|error| error.to_string())?;
        Ok(Self {
            client,
            api_key: api_key.filter(|value| !value.trim().is_empty()),
            cve_api: reqwest::Url::parse(cve_api).map_err(|error| error.to_string())?,
            cpe_api: reqwest::Url::parse(cpe_api).map_err(|error| error.to_string())?,
        })
    }

    async fn cpe_exists(&self, cpe_name: &str) -> Result<bool, String> {
        let mut request = self
            .client
            .get(self.cpe_api.clone())
            .query(&[("cpeMatchString", cpe_name), ("resultsPerPage", "10")]);
        if let Some(api_key) = &self.api_key {
            request = request.header("apiKey", api_key);
        }
        let response = request.send().await.map_err(|error| error.to_string())?;
        if !response.status().is_success() {
            return Err(format!("NVD CPE API returned {}", response.status()));
        }
        let body = response.bytes().await.map_err(|error| error.to_string())?;
        if body.len() > MAX_RESPONSE_BYTES {
            return Err("NVD CPE response exceeds limit".to_owned());
        }
        let value: serde_json::Value =
            serde_json::from_slice(&body).map_err(|error| error.to_string())?;
        let products = value["products"]
            .as_array()
            .ok_or_else(|| "NVD CPE response has no products array".to_owned())?;
        Ok(products
            .iter()
            .any(|product| product["cpe"]["cpeName"].as_str() == Some(cpe_name)))
    }
}

impl SocketRegistrationProbe {
    pub fn new(socket_path: impl Into<PathBuf>) -> Self {
        Self {
            socket_path: socket_path.into(),
        }
    }
}

impl HttpRegistrationProbe {
    pub fn new(endpoint: &str, token: impl Into<String>) -> Result<Self, String> {
        let endpoint = reqwest::Url::parse(endpoint).map_err(|error| error.to_string())?;
        if endpoint.scheme() != "https" {
            return Err("registration provider endpoint must use HTTPS".to_owned());
        }
        Self::with_endpoint(endpoint, token)
    }

    #[cfg(test)]
    fn for_test(endpoint: &str, token: impl Into<String>) -> Result<Self, String> {
        let endpoint = reqwest::Url::parse(endpoint).map_err(|error| error.to_string())?;
        Self::with_endpoint(endpoint, token)
    }

    fn with_endpoint(endpoint: reqwest::Url, token: impl Into<String>) -> Result<Self, String> {
        let token = token.into();
        if token.trim().is_empty() {
            return Err("registration provider token is required".to_owned());
        }
        let _ = rustls::crypto::ring::default_provider().install_default();
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(30))
            .redirect(reqwest::redirect::Policy::none())
            .user_agent("CyberEdge/0.1 registration-intelligence")
            .build()
            .map_err(|error| error.to_string())?;
        Ok(Self {
            client,
            endpoint,
            token,
        })
    }
}

#[derive(Serialize)]
struct SearchRequest<'a> {
    domains: &'a [String],
}

#[derive(Serialize)]
struct CveRequest<'a> {
    cpe_names: &'a [String],
}

#[derive(Serialize)]
struct RegistrationRequest<'a> {
    domains: &'a [String],
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct RegistrationResponse {
    coverage: Vec<RegistrationCoverage>,
    records: Vec<RegistrationRecord>,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct RegistrationCoverage {
    domain: String,
    status: String,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct RegistrationRecord {
    domain: String,
    icp_number: String,
    site_name: String,
    entity_name: String,
    entity_type: String,
    unified_social_credit_code: String,
    status: String,
    approved_at: String,
    source: String,
    source_url: String,
    #[serde(default)]
    relationships: Vec<RegistrationRelationship>,
}

#[derive(Deserialize, Serialize)]
#[serde(deny_unknown_fields)]
struct RegistrationRelationship {
    relationship_type: String,
    related_entity: String,
    related_domain: String,
    confidence: String,
    source: String,
}

#[derive(Deserialize)]
struct GitHubSearchResponse {
    items: Vec<GitHubCodeItem>,
}

#[derive(Deserialize)]
struct GitHubCodeItem {
    name: String,
    path: String,
    sha: String,
    html_url: String,
    repository: GitHubRepository,
}

#[derive(Deserialize)]
struct GitHubRepository {
    full_name: String,
    private: bool,
}

#[derive(Serialize)]
struct PublicCodeReference {
    query_domain: String,
    repository: String,
    path: String,
    name: String,
    blob_sha: String,
    html_url: String,
}

#[async_trait]
impl PublicCodeProbe for SocketPublicCodeProbe {
    async fn search(&self, domains: &[String]) -> Result<Vec<u8>, String> {
        validate_domains(domains)?;
        let request = serde_json::to_vec(&SearchRequest { domains })
            .map_err(|error| format!("encode public code request: {error}"))?;
        if request.len() > MAX_REQUEST_BYTES {
            return Err("public code request exceeds IPC limit".to_owned());
        }
        let exchange = async {
            let mut stream = tokio::net::UnixStream::connect(&self.socket_path)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_u32(request.len() as u32)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_all(&request)
                .await
                .map_err(|error| error.to_string())?;
            let status = stream.read_u8().await.map_err(|error| error.to_string())?;
            let length = stream.read_u32().await.map_err(|error| error.to_string())? as usize;
            if length > MAX_RESPONSE_BYTES {
                return Err("public code response exceeds IPC limit".to_owned());
            }
            let mut payload = vec![0; length];
            stream
                .read_exact(&mut payload)
                .await
                .map_err(|error| error.to_string())?;
            if status == 0 {
                Ok(payload)
            } else {
                Err(String::from_utf8_lossy(&payload).into_owned())
            }
        };
        tokio::time::timeout(std::time::Duration::from_secs(3 * 60), exchange)
            .await
            .map_err(|_| "public code adapter IPC timed out".to_owned())?
    }
}

#[async_trait]
impl PublicCodeProbe for GitHubPublicCodeProbe {
    async fn search(&self, domains: &[String]) -> Result<Vec<u8>, String> {
        validate_domains(domains)?;
        let endpoint = self
            .api_base
            .join("search/code")
            .map_err(|error| error.to_string())?;
        let mut normalized = Vec::new();
        for domain in domains {
            let query = format!("\"{domain}\"");
            let response = self
                .client
                .get(endpoint.clone())
                .bearer_auth(&self.token)
                .header("Accept", "application/vnd.github+json")
                .header("X-GitHub-Api-Version", "2026-03-10")
                .query(&[("q", query.as_str()), ("per_page", "20")])
                .send()
                .await
                .map_err(|error| error.to_string())?;
            if !response.status().is_success() {
                return Err(format!("GitHub code search returned {}", response.status()));
            }
            let body = response.bytes().await.map_err(|error| error.to_string())?;
            if body.len() > MAX_RESPONSE_BYTES {
                return Err("GitHub code search response exceeds limit".to_owned());
            }
            let search: GitHubSearchResponse =
                serde_json::from_slice(&body).map_err(|error| error.to_string())?;
            for item in search
                .items
                .into_iter()
                .filter(|item| !item.repository.private)
            {
                normalized.push(PublicCodeReference {
                    query_domain: domain.clone(),
                    repository: item.repository.full_name,
                    path: item.path,
                    name: item.name,
                    blob_sha: item.sha,
                    html_url: item.html_url,
                });
            }
        }
        serde_json::to_vec(&normalized).map_err(|error| error.to_string())
    }
}

#[async_trait]
impl CveProbe for SocketCveProbe {
    async fn query(&self, cpe_names: &[String]) -> Result<Vec<u8>, String> {
        validate_cpe_names(cpe_names)?;
        let request = serde_json::to_vec(&CveRequest { cpe_names })
            .map_err(|error| format!("encode CVE request: {error}"))?;
        if request.len() > MAX_REQUEST_BYTES {
            return Err("CVE request exceeds IPC limit".to_owned());
        }
        let exchange = async {
            let mut stream = tokio::net::UnixStream::connect(&self.socket_path)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_u32(request.len() as u32)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_all(&request)
                .await
                .map_err(|error| error.to_string())?;
            let status = stream.read_u8().await.map_err(|error| error.to_string())?;
            let length = stream.read_u32().await.map_err(|error| error.to_string())? as usize;
            if length > MAX_CVE_RESPONSE_BYTES {
                return Err("CVE response exceeds IPC limit".to_owned());
            }
            let mut payload = vec![0; length];
            stream
                .read_exact(&mut payload)
                .await
                .map_err(|error| error.to_string())?;
            if status == 0 {
                Ok(payload)
            } else {
                Err(String::from_utf8_lossy(&payload).into_owned())
            }
        };
        tokio::time::timeout(std::time::Duration::from_secs(5 * 60), exchange)
            .await
            .map_err(|_| "CVE adapter IPC timed out".to_owned())?
    }
}

#[async_trait]
impl CveProbe for NvdCveProbe {
    async fn query(&self, cpe_names: &[String]) -> Result<Vec<u8>, String> {
        validate_cpe_names(cpe_names)?;
        let mut normalized = Vec::new();
        for cpe_name in cpe_names {
            if !self.cpe_exists(cpe_name).await? {
                continue;
            }
            let mut request = self
                .client
                .get(self.cve_api.clone())
                .query(&[("cpeName", cpe_name.as_str()), ("resultsPerPage", "2000")]);
            if let Some(api_key) = &self.api_key {
                request = request.header("apiKey", api_key);
            }
            let response = request.send().await.map_err(|error| error.to_string())?;
            if !response.status().is_success() {
                return Err(format!("NVD CVE API returned {}", response.status()));
            }
            let body = response.bytes().await.map_err(|error| error.to_string())?;
            if body.len() > MAX_CVE_RESPONSE_BYTES {
                return Err("NVD CVE response exceeds limit".to_owned());
            }
            let value: serde_json::Value =
                serde_json::from_slice(&body).map_err(|error| error.to_string())?;
            let vulnerabilities = value["vulnerabilities"]
                .as_array()
                .ok_or_else(|| "NVD CVE response has no vulnerabilities array".to_owned())?;
            for vulnerability in vulnerabilities {
                normalized.push(normalize_nvd_cve(cpe_name, &vulnerability["cve"])?);
            }
        }
        let output = serde_json::to_vec(&normalized).map_err(|error| error.to_string())?;
        if output.len() > MAX_CVE_RESPONSE_BYTES {
            return Err("normalized CVE response exceeds limit".to_owned());
        }
        Ok(output)
    }
}

#[async_trait]
impl RegistrationProbe for SocketRegistrationProbe {
    async fn lookup(&self, domains: &[String]) -> Result<Vec<u8>, String> {
        validate_domains(domains)?;
        let request = serde_json::to_vec(&RegistrationRequest { domains })
            .map_err(|error| format!("encode registration request: {error}"))?;
        if request.len() > MAX_REQUEST_BYTES {
            return Err("registration request exceeds IPC limit".to_owned());
        }
        let exchange = async {
            let mut stream = tokio::net::UnixStream::connect(&self.socket_path)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_u32(request.len() as u32)
                .await
                .map_err(|error| error.to_string())?;
            stream
                .write_all(&request)
                .await
                .map_err(|error| error.to_string())?;
            let status = stream.read_u8().await.map_err(|error| error.to_string())?;
            let length = stream.read_u32().await.map_err(|error| error.to_string())? as usize;
            if length > MAX_REGISTRATION_RESPONSE_BYTES {
                return Err("registration response exceeds IPC limit".to_owned());
            }
            let mut payload = vec![0; length];
            stream
                .read_exact(&mut payload)
                .await
                .map_err(|error| error.to_string())?;
            if status == 0 {
                Ok(payload)
            } else {
                Err(String::from_utf8_lossy(&payload).into_owned())
            }
        };
        tokio::time::timeout(std::time::Duration::from_secs(5 * 60), exchange)
            .await
            .map_err(|_| "registration adapter IPC timed out".to_owned())?
    }
}

#[async_trait]
impl RegistrationProbe for HttpRegistrationProbe {
    async fn lookup(&self, domains: &[String]) -> Result<Vec<u8>, String> {
        validate_domains(domains)?;
        let response = self
            .client
            .post(self.endpoint.clone())
            .bearer_auth(&self.token)
            .json(&RegistrationRequest { domains })
            .send()
            .await
            .map_err(|error| error.to_string())?;
        if !response.status().is_success() {
            return Err(format!(
                "registration provider returned {}",
                response.status()
            ));
        }
        let body = response.bytes().await.map_err(|error| error.to_string())?;
        if body.len() > MAX_REGISTRATION_RESPONSE_BYTES {
            return Err("registration provider response exceeds limit".to_owned());
        }
        let mut response: RegistrationResponse =
            serde_json::from_slice(&body).map_err(|error| error.to_string())?;
        validate_registration_response(domains, &mut response)?;
        serde_json::to_vec(&response).map_err(|error| error.to_string())
    }
}

fn validate_registration_response(
    domains: &[String],
    response: &mut RegistrationResponse,
) -> Result<(), String> {
    if response.records.len() > 160 || response.coverage.len() != domains.len() {
        return Err("registration provider result count is invalid".to_owned());
    }
    let requested = domains.iter().collect::<std::collections::BTreeSet<_>>();
    let covered = response
        .coverage
        .iter()
        .filter(|coverage| coverage.status == "complete")
        .map(|coverage| &coverage.domain)
        .collect::<std::collections::BTreeSet<_>>();
    if requested != covered {
        return Err("registration provider coverage is incomplete".to_owned());
    }
    for record in &mut response.records {
        if !requested.contains(&record.domain)
            || record.icp_number.is_empty()
            || record.icp_number.len() > 128
            || record.site_name.len() > 512
            || record.entity_name.len() > 512
            || !matches!(
                record.entity_type.as_str(),
                "enterprise" | "government" | "institution" | "individual" | "other"
            )
            || !matches!(record.status.as_str(), "active" | "cancelled" | "unknown")
            || record.approved_at.len() > 64
            || record.source.is_empty()
            || record.source.len() > 256
            || !valid_source_url(&record.source_url)
            || !record.unified_social_credit_code.is_empty()
                && (record.unified_social_credit_code.len() != 18
                    || !record
                        .unified_social_credit_code
                        .bytes()
                        .all(|byte| byte.is_ascii_uppercase() || byte.is_ascii_digit()))
            || record.relationships.len() > 20
        {
            return Err("registration provider record violates the adapter profile".to_owned());
        }
        if record.entity_type == "individual" {
            record.entity_name.clear();
            record.unified_social_credit_code.clear();
        }
        for relationship in &record.relationships {
            if !matches!(
                relationship.relationship_type.as_str(),
                "same_entity" | "parent" | "subsidiary" | "affiliate"
            ) || relationship.related_entity.len() > 512
                || !relationship.related_domain.is_empty()
                    && validate_domains(std::slice::from_ref(&relationship.related_domain)).is_err()
                || !matches!(relationship.confidence.as_str(), "confirmed" | "reported")
                || relationship.source.is_empty()
                || relationship.source.len() > 256
            {
                return Err("registration relationship violates the adapter profile".to_owned());
            }
        }
        record.relationships.sort_by(|left, right| {
            (
                &left.relationship_type,
                &left.related_entity,
                &left.related_domain,
            )
                .cmp(&(
                    &right.relationship_type,
                    &right.related_entity,
                    &right.related_domain,
                ))
        });
    }
    response
        .coverage
        .sort_by(|left, right| left.domain.cmp(&right.domain));
    response.records.sort_by(|left, right| {
        (&left.domain, &left.icp_number, &left.entity_name).cmp(&(
            &right.domain,
            &right.icp_number,
            &right.entity_name,
        ))
    });
    Ok(())
}

fn valid_source_url(value: &str) -> bool {
    reqwest::Url::parse(value).is_ok_and(|url| {
        url.scheme() == "https"
            && url.host().is_some()
            && url.username().is_empty()
            && url.password().is_none()
    })
}

fn normalize_nvd_cve(cpe_name: &str, cve: &serde_json::Value) -> Result<serde_json::Value, String> {
    let id = cve["id"]
        .as_str()
        .filter(|id| valid_cve_id(id))
        .ok_or_else(|| "NVD CVE record has invalid ID".to_owned())?;
    let description = cve["descriptions"]
        .as_array()
        .and_then(|values| values.iter().find(|value| value["lang"] == "en"))
        .and_then(|value| value["value"].as_str())
        .unwrap_or_default()
        .chars()
        .take(4096)
        .collect::<String>();
    let metric = [
        "cvssMetricV40",
        "cvssMetricV31",
        "cvssMetricV30",
        "cvssMetricV2",
    ]
    .into_iter()
    .find_map(|name| {
        cve["metrics"][name]
            .as_array()
            .and_then(|values| values.first())
    });
    let cvss_data = metric.map(|value| &value["cvssData"]);
    let base_severity = cvss_data
        .and_then(|value| value["baseSeverity"].as_str())
        .or_else(|| metric.and_then(|value| value["baseSeverity"].as_str()))
        .unwrap_or("UNKNOWN");
    let references = cve["references"]
        .as_array()
        .into_iter()
        .flatten()
        .filter_map(|value| value["url"].as_str())
        .filter(|url| url.starts_with("https://") || url.starts_with("http://"))
        .take(10)
        .collect::<Vec<_>>();
    Ok(serde_json::json!({
        "cpe_name": cpe_name,
        "cve_id": id,
        "source_identifier": cve["sourceIdentifier"].as_str().unwrap_or_default(),
        "published": cve["published"].as_str().unwrap_or_default(),
        "last_modified": cve["lastModified"].as_str().unwrap_or_default(),
        "vuln_status": cve["vulnStatus"].as_str().unwrap_or_default(),
        "description": description,
        "cvss_version": cvss_data.and_then(|value| value["version"].as_str()).unwrap_or_default(),
        "cvss_vector": cvss_data.and_then(|value| value["vectorString"].as_str()).unwrap_or_default(),
        "base_score": cvss_data.and_then(|value| value["baseScore"].as_f64()),
        "base_severity": base_severity,
        "references": references,
    }))
}

fn validate_domains(domains: &[String]) -> Result<(), String> {
    if domains.is_empty() || domains.len() > MAX_DOMAINS {
        return Err(format!(
            "public code search requires 1..={MAX_DOMAINS} domains"
        ));
    }
    for domain in domains {
        if domain.len() > MAX_DOMAIN_BYTES
            || domain.is_empty()
            || domain.starts_with('.')
            || domain.ends_with('.')
            || !domain
                .bytes()
                .all(|byte| byte.is_ascii_alphanumeric() || matches!(byte, b'.' | b'-'))
        {
            return Err("public code search domain is invalid".to_owned());
        }
    }
    Ok(())
}

fn validate_cpe_names(cpe_names: &[String]) -> Result<(), String> {
    if cpe_names.is_empty() || cpe_names.len() > MAX_CPE_NAMES {
        return Err(format!("CVE query requires 1..={MAX_CPE_NAMES} CPE names"));
    }
    for cpe_name in cpe_names {
        let components = cpe_name.split(':').collect::<Vec<_>>();
        if components.len() != 13
            || components[0] != "cpe"
            || components[1] != "2.3"
            || !matches!(components[2], "a" | "o" | "h")
            || components[3..=5]
                .iter()
                .any(|value| value.is_empty() || matches!(*value, "*" | "-"))
            || cpe_name.len() > 1024
            || cpe_name.chars().any(char::is_control)
        {
            return Err("CVE query CPE name is not exact".to_owned());
        }
    }
    Ok(())
}

fn valid_cve_id(value: &str) -> bool {
    let mut parts = value.split('-');
    parts.next() == Some("CVE")
        && parts
            .next()
            .is_some_and(|year| year.len() == 4 && year.bytes().all(|b| b.is_ascii_digit()))
        && parts
            .next()
            .is_some_and(|number| number.len() >= 4 && number.bytes().all(|b| b.is_ascii_digit()))
        && parts.next().is_none()
}

#[cfg(test)]
mod tests {
    use super::{
        CveProbe, GitHubPublicCodeProbe, HttpRegistrationProbe, NvdCveProbe, PublicCodeProbe,
        RegistrationProbe, validate_cpe_names, validate_domains,
    };
    use tokio::{
        io::{AsyncReadExt, AsyncWriteExt},
        net::TcpListener,
    };

    #[test]
    fn accepts_only_bounded_dns_names() {
        assert!(validate_domains(&["example.com".to_owned()]).is_ok());
        assert!(validate_domains(&[]).is_err());
        assert!(validate_domains(&["https://example.com".to_owned()]).is_err());
        assert!(validate_domains(&["example.com OR org:other".to_owned()]).is_err());
    }

    #[test]
    fn accepts_only_exact_cpe_names() {
        assert!(
            validate_cpe_names(&["cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*".to_owned()])
                .is_ok()
        );
        assert!(
            validate_cpe_names(&["cpe:2.3:a:wordpress:wordpress:*:*:*:*:*:*:*:*".to_owned()])
                .is_err()
        );
        assert!(validate_cpe_names(&[]).is_err());
    }

    #[tokio::test]
    async fn github_probe_returns_only_public_reference_metadata() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();
        let server = tokio::spawn(async move {
            let (mut stream, _) = listener.accept().await.unwrap();
            let mut request = vec![0; 4096];
            let length = stream.read(&mut request).await.unwrap();
            let request = String::from_utf8_lossy(&request[..length]);
            assert!(request.starts_with("GET /search/code?"));
            assert!(request.contains("q=%22example.com%22"));
            assert!(request.contains("per_page=20"));
            assert!(request.contains("authorization: Bearer test-token"));
            assert!(request.contains("x-github-api-version: 2026-03-10"));
            let body = serde_json::json!({"items": [{
                "name": "config.yaml",
                "path": "deploy/config.yaml",
                "sha": "0123456789abcdef0123456789abcdef01234567",
                "html_url": "https://github.com/example/repo/blob/main/deploy/config.yaml",
                "repository": {"full_name": "example/repo", "private": false},
                "text_matches": [{"fragment": "must never be retained"}]
            }, {
                "name": "private.txt",
                "path": "private.txt",
                "sha": "abcdef0123456789abcdef0123456789abcdef01",
                "html_url": "https://github.com/example/private/blob/main/private.txt",
                "repository": {"full_name": "example/private", "private": true}
            }]})
            .to_string();
            let response = format!(
                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(response.as_bytes()).await.unwrap();
        });
        let probe =
            GitHubPublicCodeProbe::with_api_base("test-token", &format!("http://{address}/"))
                .unwrap();
        let output = probe.search(&["example.com".to_owned()]).await.unwrap();
        let output: serde_json::Value = serde_json::from_slice(&output).unwrap();
        assert_eq!(output.as_array().unwrap().len(), 1);
        assert_eq!(output[0]["repository"], "example/repo");
        assert!(output[0].get("text_matches").is_none());
        assert!(!output.to_string().contains("must never be retained"));
        server.await.unwrap();
    }

    #[tokio::test]
    async fn nvd_probe_normalizes_exact_cpe_results() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();
        let server = tokio::spawn(async move {
            for index in 0..2 {
                let (mut stream, _) = listener.accept().await.unwrap();
                let mut request = vec![0; 4096];
                let length = stream.read(&mut request).await.unwrap();
                let request = String::from_utf8_lossy(&request[..length]);
                assert!(request.starts_with("GET /?"));
                assert!(request.contains("apikey: test-key"));
                let body = if index == 0 {
                    assert!(request.contains("cpeMatchString=cpe%3A2.3%3Aa%3Awordpress"));
                    assert!(request.contains("resultsPerPage=10"));
                    serde_json::json!({"products": [{"cpe": {
                        "cpeName": "cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*"
                    }}]})
                } else {
                    assert!(request.contains("cpeName=cpe%3A2.3%3Aa%3Awordpress"));
                    assert!(request.contains("resultsPerPage=2000"));
                    serde_json::json!({"vulnerabilities": [{"cve": {
                        "id": "CVE-2099-1234",
                        "sourceIdentifier": "security@example.test",
                        "published": "2099-01-01T00:00:00.000",
                        "lastModified": "2099-01-02T00:00:00.000",
                        "vulnStatus": "Analyzed",
                        "descriptions": [{"lang": "en", "value": "Evidence-backed test CVE."}],
                        "metrics": {"cvssMetricV31": [{"cvssData": {
                            "version": "3.1", "vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
                            "baseScore": 9.8, "baseSeverity": "CRITICAL"
                        }}]},
                        "references": [{"url": "https://example.test/CVE-2099-1234"}],
                        "configurations": [{"must_not_be_retained": "raw applicability tree"}]
                    }}]})
                }
                .to_string();
                let response = format!(
                    "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                    body.len(),
                    body
                );
                stream.write_all(response.as_bytes()).await.unwrap();
            }
        });
        let probe = NvdCveProbe::with_api_bases(
            Some("test-key".to_owned()),
            &format!("http://{address}/"),
            &format!("http://{address}/"),
        )
        .unwrap();
        let cpe = "cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*".to_owned();
        let output = probe.query(std::slice::from_ref(&cpe)).await.unwrap();
        let output: serde_json::Value = serde_json::from_slice(&output).unwrap();
        assert_eq!(output[0]["cpe_name"], cpe);
        assert_eq!(output[0]["cve_id"], "CVE-2099-1234");
        assert_eq!(output[0]["base_severity"], "CRITICAL");
        assert!(output[0].get("configurations").is_none());
        assert!(!output.to_string().contains("must_not_be_retained"));
        server.await.unwrap();
    }

    #[tokio::test]
    async fn registration_probe_requires_coverage_and_redacts_individuals() {
        let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
        let address = listener.local_addr().unwrap();
        let server = tokio::spawn(async move {
            let (mut stream, _) = listener.accept().await.unwrap();
            let mut request = vec![0; 4096];
            let length = stream.read(&mut request).await.unwrap();
            let request = String::from_utf8_lossy(&request[..length]);
            assert!(request.starts_with("POST /lookup "));
            assert!(request.contains("authorization: Bearer provider-token"));
            assert!(request.contains("example.cn"));
            let body = serde_json::json!({
                "coverage": [{"domain": "example.cn", "status": "complete"}],
                "records": [{
                    "domain": "example.cn",
                    "icp_number": "京ICP备00000000号-1",
                    "site_name": "Example",
                    "entity_name": "must be redacted",
                    "entity_type": "individual",
                    "unified_social_credit_code": "",
                    "status": "active",
                    "approved_at": "2026-01-01",
                    "source": "licensed-provider",
                    "source_url": "https://provider.example/records/1",
                    "relationships": []
                }]
            })
            .to_string();
            let response = format!(
                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            stream.write_all(response.as_bytes()).await.unwrap();
        });
        let probe =
            HttpRegistrationProbe::for_test(&format!("http://{address}/lookup"), "provider-token")
                .unwrap();
        let output = probe.lookup(&["example.cn".to_owned()]).await.unwrap();
        let output: serde_json::Value = serde_json::from_slice(&output).unwrap();
        assert_eq!(output["records"][0]["entity_name"], "");
        assert!(!output.to_string().contains("must be redacted"));
        server.await.unwrap();
    }
}
