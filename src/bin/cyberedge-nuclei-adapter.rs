use std::{
    env, io,
    os::unix::fs::{FileTypeExt, PermissionsExt},
    path::Path,
};

use cyberedge::{NucleiProbe, SystemNucleiProbe};
use serde::Deserialize;
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::{UnixListener, UnixStream},
};

const MAX_REQUEST_BYTES: usize = 128 * 1024;
const MAX_ERROR_BYTES: usize = 4096;

#[derive(Deserialize)]
struct ScanRequest {
    targets: Vec<String>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let socket = env::var("CYBEREDGE_NUCLEI_SOCKET")
        .unwrap_or_else(|_| "/run/cyberedge-nuclei/adapter.sock".to_owned());
    let binary =
        env::var("CYBEREDGE_NUCLEI_BIN").unwrap_or_else(|_| "/usr/local/bin/nuclei".to_owned());
    let templates = env::var("CYBEREDGE_NUCLEI_TEMPLATES")
        .unwrap_or_else(|_| "/opt/cyberedge/nuclei-templates".to_owned());
    prepare_socket(Path::new(&socket))?;
    let listener = UnixListener::bind(&socket)?;
    std::fs::set_permissions(&socket, std::fs::Permissions::from_mode(0o600))?;
    let probe = SystemNucleiProbe::new(binary, templates);
    eprintln!("CyberEdge Nuclei adapter listening on unix://{socket}");
    loop {
        let (stream, _) = listener.accept().await?;
        if let Err(error) = handle(stream, &probe).await {
            eprintln!("nuclei adapter request error: {error}");
        }
    }
}

async fn handle(mut stream: UnixStream, probe: &SystemNucleiProbe) -> Result<(), io::Error> {
    let length = stream.read_u32().await? as usize;
    if length > MAX_REQUEST_BYTES {
        return write_error(&mut stream, "nuclei request exceeds adapter limit").await;
    }
    let mut request = vec![0; length];
    stream.read_exact(&mut request).await?;
    let request: ScanRequest = match serde_json::from_slice(&request) {
        Ok(request) => request,
        Err(error) => return write_error(&mut stream, &format!("invalid request: {error}")).await,
    };
    match probe.scan(&request.targets).await {
        Ok(output) => {
            stream.write_u8(0).await?;
            stream.write_u32(output.len() as u32).await?;
            stream.write_all(&output).await?;
            stream.flush().await
        }
        Err(error) => write_error(&mut stream, &error).await,
    }
}

async fn write_error(stream: &mut UnixStream, error: &str) -> Result<(), io::Error> {
    let bytes = error.as_bytes();
    let bytes = &bytes[..bytes.len().min(MAX_ERROR_BYTES)];
    stream.write_u8(1).await?;
    stream.write_u32(bytes.len() as u32).await?;
    stream.write_all(bytes).await?;
    stream.flush().await
}

fn prepare_socket(path: &Path) -> Result<(), io::Error> {
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    match std::fs::symlink_metadata(path) {
        Ok(metadata) if metadata.file_type().is_socket() => std::fs::remove_file(path),
        Ok(_) => Err(io::Error::new(
            io::ErrorKind::AlreadyExists,
            "nuclei socket path exists and is not a socket",
        )),
        Err(error) if error.kind() == io::ErrorKind::NotFound => Ok(()),
        Err(error) => Err(error),
    }
}
