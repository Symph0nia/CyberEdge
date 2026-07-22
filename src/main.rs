use std::{env, io, os::unix::fs::FileTypeExt, path::Path, sync::Arc};

use cyberedge::{CyberEdgeService, PostgresRepository, StaticAuthorizer};
use tokio::{net::UnixListener, signal};
use tokio_stream::wrappers::UnixListenerStream;
use tonic::transport::Server;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let database_url = env::var("DATABASE_URL").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "DATABASE_URL is required for persistent runtime",
        )
    })?;
    let socket_path =
        env::var("CYBEREDGE_RPC_SOCKET").unwrap_or_else(|_| "/tmp/cyberedge.sock".to_owned());
    let policy_path = env::var("CYBEREDGE_AGENT_POLICY").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "CYBEREDGE_AGENT_POLICY is required",
        )
    })?;
    prepare_socket(Path::new(&socket_path))?;

    let repository = Arc::new(PostgresRepository::connect(&database_url).await?);
    let authorizer = Arc::new(StaticAuthorizer::load(policy_path)?);
    let listener = UnixListener::bind(&socket_path)?;
    println!("CyberEdge RPC listening on unix://{socket_path}");
    Server::builder()
        .add_service(CyberEdgeService::new(repository, authorizer).server())
        .serve_with_incoming_shutdown(UnixListenerStream::new(listener), shutdown_signal())
        .await?;

    Ok(())
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
