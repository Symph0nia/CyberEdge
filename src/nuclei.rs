use std::{path::PathBuf, process::Stdio};

use async_trait::async_trait;
use serde::Serialize;
use tokio::io::{AsyncReadExt, AsyncWriteExt};

const MAX_TARGETS: usize = 64;
const MAX_TARGET_BYTES: usize = 2048;
const MAX_REQUEST_BYTES: usize = 128 * 1024;
const MAX_RESPONSE_BYTES: usize = 10 * 1024 * 1024;

#[async_trait]
pub trait NucleiProbe: Send + Sync {
    async fn scan(&self, targets: &[String]) -> Result<Vec<u8>, String>;
}

pub struct SocketNucleiProbe {
    socket_path: PathBuf,
}

pub struct SystemNucleiProbe {
    binary: PathBuf,
    templates: PathBuf,
}

impl SystemNucleiProbe {
    pub fn new(binary: impl Into<PathBuf>, templates: impl Into<PathBuf>) -> Self {
        Self {
            binary: binary.into(),
            templates: templates.into(),
        }
    }
}

impl SocketNucleiProbe {
    pub fn new(socket_path: impl Into<PathBuf>) -> Self {
        Self {
            socket_path: socket_path.into(),
        }
    }
}

#[derive(Serialize)]
struct ScanRequest<'a> {
    targets: &'a [String],
}

#[async_trait]
impl NucleiProbe for SocketNucleiProbe {
    async fn scan(&self, targets: &[String]) -> Result<Vec<u8>, String> {
        validate_targets(targets)?;
        let request = serde_json::to_vec(&ScanRequest { targets })
            .map_err(|error| format!("encode nuclei request: {error}"))?;
        if request.len() > MAX_REQUEST_BYTES {
            return Err("nuclei request exceeds IPC limit".to_owned());
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
            stream.flush().await.map_err(|error| error.to_string())?;
            let status = stream.read_u8().await.map_err(|error| error.to_string())?;
            let length = stream.read_u32().await.map_err(|error| error.to_string())? as usize;
            if length > MAX_RESPONSE_BYTES {
                return Err("nuclei response exceeds IPC limit".to_owned());
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
        tokio::time::timeout(std::time::Duration::from_secs(10 * 60), exchange)
            .await
            .map_err(|_| "nuclei adapter IPC timed out".to_owned())?
    }
}

#[async_trait]
impl NucleiProbe for SystemNucleiProbe {
    async fn scan(&self, targets: &[String]) -> Result<Vec<u8>, String> {
        validate_targets(targets)?;
        if !self.binary.is_absolute() || !self.templates.is_absolute() {
            return Err("nuclei binary and template paths must be absolute".to_owned());
        }
        let metadata = tokio::fs::metadata(&self.templates)
            .await
            .map_err(|error| format!("nuclei templates unavailable: {error}"))?;
        if !metadata.is_dir() {
            return Err("nuclei template allowlist is not a directory".to_owned());
        }
        let token = uuid::Uuid::now_v7();
        let work = std::env::temp_dir().join(format!("cyberedge-nuclei-{token}"));
        let target_file = work.join("targets.txt");
        tokio::fs::create_dir(&work)
            .await
            .map_err(|error| error.to_string())?;
        let mut target_content = targets.join("\n");
        target_content.push('\n');
        if let Err(error) = tokio::fs::write(&target_file, target_content).await {
            let _ = tokio::fs::remove_dir(&work).await;
            return Err(error.to_string());
        }
        let command = tokio::process::Command::new(&self.binary)
            .kill_on_drop(true)
            .env_clear()
            .env("HOME", &work)
            .env("PATH", "/usr/local/bin:/usr/bin:/bin")
            .env("DISABLE_NUCLEI_TEMPLATES_PUBLIC_DOWNLOAD", "true")
            .env("DISABLE_NUCLEI_TEMPLATES_GITHUB_DOWNLOAD", "true")
            .env("DISABLE_NUCLEI_TEMPLATES_GITLAB_DOWNLOAD", "true")
            .env("DISABLE_NUCLEI_TEMPLATES_AWS_DOWNLOAD", "true")
            .env("DISABLE_NUCLEI_TEMPLATES_AZURE_DOWNLOAD", "true")
            .stdin(Stdio::null())
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .args([
                "-list",
                target_file.to_string_lossy().as_ref(),
                "-templates",
                self.templates.to_string_lossy().as_ref(),
                "-type",
                "http,ssl",
                "-severity",
                "info,low,medium,high,critical",
                "-exclude-tags",
                "fuzz,dos,headless,code",
                "-disable-unsigned-templates",
                "-disable-update-check",
                "-no-interactsh",
                "-no-stdin",
                "-disable-redirects",
                "-jsonl",
                "-silent",
                "-no-color",
                "-omit-raw",
                "-omit-template",
                "-rate-limit",
                "20",
                "-bulk-size",
                "5",
                "-concurrency",
                "5",
                "-timeout",
                "5",
                "-retries",
                "1",
                "-max-host-error",
                "10",
            ])
            .output();
        let result = tokio::time::timeout(std::time::Duration::from_secs(9 * 60), command)
            .await
            .map_err(|_| "nuclei scan timed out".to_owned())
            .and_then(|result| result.map_err(|error| error.to_string()));
        let _ = tokio::fs::remove_file(&target_file).await;
        let _ = tokio::fs::remove_dir_all(&work).await;
        let output = result?;
        if !output.status.success() {
            let error = String::from_utf8_lossy(&output.stderr);
            return Err(format!(
                "nuclei exited with {}: {}",
                output.status,
                error.chars().take(4096).collect::<String>()
            ));
        }
        if output.stdout.len() > MAX_RESPONSE_BYTES {
            return Err("nuclei output exceeds evidence limit".to_owned());
        }
        Ok(output.stdout)
    }
}

fn validate_targets(targets: &[String]) -> Result<(), String> {
    if targets.is_empty() || targets.len() > MAX_TARGETS {
        return Err(format!("nuclei target count must be 1..={MAX_TARGETS}"));
    }
    for target in targets {
        if target.len() > MAX_TARGET_BYTES {
            return Err("nuclei target exceeds URL limit".to_owned());
        }
        let url = reqwest::Url::parse(target).map_err(|_| "invalid nuclei target URL")?;
        if !matches!(url.scheme(), "http" | "https")
            || !url.username().is_empty()
            || url.password().is_some()
            || url.host().is_none()
        {
            return Err("nuclei targets require credential-free HTTP(S) URLs".to_owned());
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::os::unix::fs::PermissionsExt;

    #[test]
    fn accepts_only_bounded_http_targets_without_credentials() {
        assert!(validate_targets(&["https://example.com/".to_owned()]).is_ok());
        assert!(validate_targets(&[]).is_err());
        assert!(validate_targets(&["file:///etc/passwd".to_owned()]).is_err());
        assert!(validate_targets(&["https://user:secret@example.com/".to_owned()]).is_err());
        assert!(
            validate_targets(
                &(0..=MAX_TARGETS)
                    .map(|index| format!("https://example.com/{index}"))
                    .collect::<Vec<_>>()
            )
            .is_err()
        );
    }

    #[tokio::test]
    async fn system_probe_enforces_the_reviewed_machine_profile() {
        let root =
            std::env::temp_dir().join(format!("cyberedge-nuclei-test-{}", uuid::Uuid::now_v7()));
        let templates = root.join("templates");
        let binary = root.join("nuclei");
        tokio::fs::create_dir_all(&templates).await.unwrap();
        tokio::fs::write(
            &binary,
            r#"#!/bin/sh
case " $* " in *" -disable-unsigned-templates "*) ;; *) exit 20;; esac
case " $* " in *" -disable-update-check "*) ;; *) exit 21;; esac
case " $* " in *" -no-interactsh "*) ;; *) exit 22;; esac
case " $* " in *" -type http,ssl "*) ;; *) exit 23;; esac
case " $* " in *" -rate-limit 20 "*) ;; *) exit 24;; esac
printf '%s' '{"template-id":"profile-test","type":"http","host":"https://example.com/","matched-at":"https://example.com/","info":{"severity":"low"}}'
"#,
        )
        .await
        .unwrap();
        std::fs::set_permissions(&binary, std::fs::Permissions::from_mode(0o700)).unwrap();
        let output = SystemNucleiProbe::new(&binary, &templates)
            .scan(&["https://example.com/".to_owned()])
            .await
            .unwrap();
        assert!(String::from_utf8_lossy(&output).contains("profile-test"));
        tokio::fs::remove_file(binary).await.unwrap();
        tokio::fs::remove_dir(templates).await.unwrap();
        tokio::fs::remove_dir(root).await.unwrap();
    }
}
