use std::path::PathBuf;

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use tokio::io::{AsyncReadExt, AsyncWriteExt};

const MAX_DOMAINS: usize = 8;
const MAX_DOMAIN_BYTES: usize = 253;
const MAX_REQUEST_BYTES: usize = 4 * 1024;
const MAX_RESPONSE_BYTES: usize = 1024 * 1024;

#[async_trait]
pub trait PublicCodeProbe: Send + Sync {
    async fn search(&self, domains: &[String]) -> Result<Vec<u8>, String>;
}

pub struct SocketPublicCodeProbe {
    socket_path: PathBuf,
}

pub struct GitHubPublicCodeProbe {
    client: reqwest::Client,
    token: String,
    api_base: reqwest::Url,
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

#[derive(Serialize)]
struct SearchRequest<'a> {
    domains: &'a [String],
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

#[cfg(test)]
mod tests {
    use super::{GitHubPublicCodeProbe, PublicCodeProbe, validate_domains};
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
}
