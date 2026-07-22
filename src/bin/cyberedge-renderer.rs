use std::{
    env, io,
    os::unix::fs::{FileTypeExt, PermissionsExt},
    path::Path,
    sync::Arc,
};

use cyberedge::{ScreenshotProbe, SystemScreenshotProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::{UnixListener, UnixStream},
};

const MAX_HTML_BYTES: usize = 1_048_576;
const MAX_ERROR_BYTES: usize = 4096;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let socket = env::var("CYBEREDGE_RENDERER_SOCKET")
        .unwrap_or_else(|_| "/run/cyberedge-renderer/render.sock".to_owned());
    let binary =
        env::var("CYBEREDGE_CHROMIUM_BIN").unwrap_or_else(|_| "/usr/bin/chromium".to_owned());
    prepare_socket(Path::new(&socket))?;
    let listener = UnixListener::bind(&socket)?;
    std::fs::set_permissions(&socket, std::fs::Permissions::from_mode(0o600))?;
    let renderer: Arc<dyn ScreenshotProbe> =
        Arc::new(SystemScreenshotProbe::isolated_container(binary));
    eprintln!("CyberEdge renderer listening on unix://{socket}");
    loop {
        let (stream, _) = listener.accept().await?;
        if let Err(error) = handle(stream, renderer.clone()).await {
            eprintln!("renderer request error: {error}");
        }
    }
}

async fn handle(
    mut stream: UnixStream,
    renderer: Arc<dyn ScreenshotProbe>,
) -> Result<(), io::Error> {
    let length = stream.read_u32().await? as usize;
    if length > MAX_HTML_BYTES {
        return write_error(&mut stream, "HTML evidence exceeds renderer limit").await;
    }
    let mut html = vec![0; length];
    stream.read_exact(&mut html).await?;
    match renderer.capture(&html).await {
        Ok(png) => {
            stream.write_u8(0).await?;
            stream.write_u32(png.len() as u32).await?;
            stream.write_all(&png).await?;
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
            "renderer socket path exists and is not a socket",
        )),
        Err(error) if error.kind() == io::ErrorKind::NotFound => Ok(()),
        Err(error) => Err(error),
    }
}
