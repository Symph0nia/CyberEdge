use std::{
    env, io,
    os::unix::fs::{FileTypeExt, PermissionsExt},
    path::Path,
};

use cyberedge::{HttpRegistrationProbe, RegistrationProbe};
use serde::Deserialize;
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::{UnixListener, UnixStream},
};

const MAX_REQUEST_BYTES: usize = 4 * 1024;
const MAX_RESPONSE_BYTES: usize = 2 * 1024 * 1024;

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct RegistrationRequest {
    domains: Vec<String>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let socket = env::var("CYBEREDGE_REGISTRATION_SOCKET")
        .unwrap_or_else(|_| "/run/cyberedge-registration/adapter.sock".to_owned());
    let endpoint = env::var("CYBEREDGE_REGISTRATION_PROVIDER_URL").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "CYBEREDGE_REGISTRATION_PROVIDER_URL is required",
        )
    })?;
    let token_path = env::var("REGISTRATION_PROVIDER_TOKEN_FILE").map_err(|_| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "REGISTRATION_PROVIDER_TOKEN_FILE is required",
        )
    })?;
    let token = std::fs::read_to_string(token_path)?.trim().to_owned();
    prepare_socket(Path::new(&socket))?;
    let listener = UnixListener::bind(&socket)?;
    std::fs::set_permissions(&socket, std::fs::Permissions::from_mode(0o600))?;
    let probe = HttpRegistrationProbe::new(&endpoint, token)?;
    loop {
        let (stream, _) = listener.accept().await?;
        if let Err(error) = handle(stream, &probe).await {
            eprintln!("registration adapter request error: {error}");
        }
    }
}

async fn handle(mut stream: UnixStream, probe: &dyn RegistrationProbe) -> Result<(), String> {
    let length = stream.read_u32().await.map_err(|error| error.to_string())? as usize;
    if length > MAX_REQUEST_BYTES {
        return write_response(&mut stream, 1, b"registration request exceeds IPC limit").await;
    }
    let mut payload = vec![0; length];
    stream
        .read_exact(&mut payload)
        .await
        .map_err(|error| error.to_string())?;
    let request: RegistrationRequest = match serde_json::from_slice(&payload) {
        Ok(request) => request,
        Err(error) => return write_response(&mut stream, 1, error.to_string().as_bytes()).await,
    };
    match probe.lookup(&request.domains).await {
        Ok(output) if output.len() <= MAX_RESPONSE_BYTES => {
            write_response(&mut stream, 0, &output).await
        }
        Ok(_) => write_response(&mut stream, 1, b"registration response exceeds IPC limit").await,
        Err(error) => write_response(&mut stream, 1, error.as_bytes()).await,
    }
}

async fn write_response(stream: &mut UnixStream, status: u8, payload: &[u8]) -> Result<(), String> {
    stream
        .write_u8(status)
        .await
        .map_err(|error| error.to_string())?;
    stream
        .write_u32(payload.len() as u32)
        .await
        .map_err(|error| error.to_string())?;
    stream
        .write_all(payload)
        .await
        .map_err(|error| error.to_string())
}

fn prepare_socket(path: &Path) -> io::Result<()> {
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    match std::fs::symlink_metadata(path) {
        Ok(metadata) if metadata.file_type().is_socket() => std::fs::remove_file(path),
        Ok(_) => Err(io::Error::new(
            io::ErrorKind::AlreadyExists,
            "registration socket path exists and is not a socket",
        )),
        Err(error) if error.kind() == io::ErrorKind::NotFound => Ok(()),
        Err(error) => Err(error),
    }
}
