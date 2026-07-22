use prost_types::Timestamp;
use std::{
    env, io,
    os::unix::fs::FileTypeExt,
    path::{Path, PathBuf},
    sync::Arc,
};

use cyberedge::{
    BASELINE_SERVICE_PORTS, CrtShSource, CyberEdgeService, DiscoveryWorker, NotificationDispatcher,
    PostgresRepository, Repository, SocketScreenshotProbe, StaticAuthorizer,
    SystemCertificateProbe, SystemDnsResolver, SystemPortConnector, SystemScreenshotProbe,
    SystemWebsiteProbe, WebhookSink, serve_read_only_web,
};
use tokio::{net::UnixListener, signal};
use tokio_stream::wrappers::UnixListenerStream;
use tonic::transport::{Certificate, Identity, Server, ServerTlsConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let database_url = env::var("DATABASE_URL").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "DATABASE_URL is required for persistent runtime",
        )
    })?;
    let policy_path = env::var("CYBEREDGE_AGENT_POLICY").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "CYBEREDGE_AGENT_POLICY is required",
        )
    })?;
    let repository = Arc::new(PostgresRepository::connect(&database_url).await?);
    let authorizer = Arc::new(StaticAuthorizer::load(policy_path)?);
    let mut discovery = DiscoveryWorker::new(repository.clone(), Arc::new(SystemDnsResolver))
        .with_certificate_source(Arc::new(CrtShSource::new()?))
        .with_port_connector(
            Arc::new(SystemPortConnector),
            BASELINE_SERVICE_PORTS.to_vec(),
        )
        .with_certificate_probe(Arc::new(SystemCertificateProbe::new()))
        .with_website_probe(Arc::new(SystemWebsiteProbe));
    if let Ok(socket) = env::var("CYBEREDGE_SCREENSHOT_RENDERER_SOCKET") {
        discovery = discovery.with_screenshot_probe(Arc::new(SocketScreenshotProbe::new(socket)));
    } else if env::var("CYBEREDGE_SCREENSHOTS_ENABLED")
        .is_ok_and(|value| value == "true" || value == "1")
    {
        let binary =
            env::var("CYBEREDGE_CHROMIUM_BIN").unwrap_or_else(|_| "/usr/bin/chromium".to_owned());
        discovery = discovery.with_screenshot_probe(Arc::new(SystemScreenshotProbe::new(binary)));
    }
    let scheduler_repository = repository.clone();
    let scheduler = tokio::spawn(async move {
        let mut interval = tokio::time::interval(std::time::Duration::from_secs(1));
        loop {
            interval.tick().await;
            let duration = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .expect("system time must be after Unix epoch");
            let timestamp = Timestamp {
                seconds: duration.as_secs() as i64,
                nanos: duration.subsec_nanos() as i32,
            };
            if let Err(error) = scheduler_repository.enqueue_due_schedules(timestamp).await {
                eprintln!("schedule worker error: {error}");
            }
        }
    });
    let worker = tokio::spawn(async move {
        let mut interval = tokio::time::interval(std::time::Duration::from_millis(500));
        loop {
            interval.tick().await;
            if let Err(error) = discovery.run_once().await {
                eprintln!("discovery worker error: {error}");
            }
        }
    });
    let notification = match env::var("CYBEREDGE_WEBHOOK_URL") {
        Ok(url) => {
            let sink = Arc::new(WebhookSink::new(
                &url,
                env::var("CYBEREDGE_WEBHOOK_BEARER_TOKEN").ok(),
            )?);
            let dispatcher = NotificationDispatcher::new(repository.clone(), sink);
            Some(tokio::spawn(async move {
                let mut interval = tokio::time::interval(std::time::Duration::from_secs(1));
                loop {
                    interval.tick().await;
                    if let Err(error) = dispatcher.dispatch_once().await {
                        eprintln!("notification dispatcher error: {error}");
                    }
                }
            }))
        }
        Err(env::VarError::NotPresent) => None,
        Err(error) => return Err(error.into()),
    };
    let web = match env::var("CYBEREDGE_WEB_BIND") {
        Ok(value) => {
            let address = value.parse()?;
            let dist = PathBuf::from(
                env::var("CYBEREDGE_WEB_DIST").unwrap_or_else(|_| "web/dist".to_owned()),
            );
            let repository = repository.clone();
            Some(tokio::spawn(async move {
                if let Err(error) = serve_read_only_web(repository, address, dist).await {
                    eprintln!("read-only web error: {error}");
                }
            }))
        }
        Err(env::VarError::NotPresent) => None,
        Err(error) => return Err(error.into()),
    };
    let service = CyberEdgeService::new(repository, authorizer).server();
    if let Ok(value) = env::var("CYBEREDGE_RPC_ADDR") {
        let address = value.parse()?;
        let identity = Identity::from_pem(
            required_file("CYBEREDGE_TLS_CERT")?,
            required_file("CYBEREDGE_TLS_KEY")?,
        );
        let client_ca = Certificate::from_pem(required_file("CYBEREDGE_TLS_CLIENT_CA")?);
        println!("CyberEdge RPC listening with mTLS on https://{address}");
        Server::builder()
            .tls_config(
                ServerTlsConfig::new()
                    .identity(identity)
                    .client_ca_root(client_ca),
            )?
            .add_service(service)
            .serve_with_shutdown(address, shutdown_signal())
            .await?;
    } else {
        let socket_path =
            env::var("CYBEREDGE_RPC_SOCKET").unwrap_or_else(|_| "/tmp/cyberedge.sock".to_owned());
        prepare_socket(Path::new(&socket_path))?;
        let listener = UnixListener::bind(&socket_path)?;
        println!("CyberEdge RPC listening on unix://{socket_path}");
        Server::builder()
            .add_service(service)
            .serve_with_incoming_shutdown(UnixListenerStream::new(listener), shutdown_signal())
            .await?;
    }
    worker.abort();
    scheduler.abort();
    if let Some(notification) = notification {
        notification.abort();
    }
    if let Some(web) = web {
        web.abort();
    }

    Ok(())
}

fn required_file(variable: &str) -> io::Result<Vec<u8>> {
    let path = env::var(variable).map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            format!("{variable} is required for mTLS"),
        )
    })?;
    std::fs::read(path)
}

fn prepare_socket(path: &Path) -> io::Result<()> {
    match path.symlink_metadata() {
        Ok(metadata) if metadata.file_type().is_socket() => std::fs::remove_file(path),
        Ok(_) => Err(io::Error::new(
            io::ErrorKind::AlreadyExists,
            "RPC socket path exists and is not a socket",
        )),
        Err(error) if error.kind() == io::ErrorKind::NotFound => Ok(()),
        Err(error) => Err(error),
    }
}

async fn shutdown_signal() {
    if let Err(error) = signal::ctrl_c().await {
        eprintln!("failed to listen for shutdown signal: {error}");
    }
}
