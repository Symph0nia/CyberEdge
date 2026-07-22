use std::path::PathBuf;

use cyberedge::{ScreenshotProbe, SocketScreenshotProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::UnixListener,
};
use uuid::Uuid;

const PNG: &[u8] = b"\x89PNG\r\n\x1a\nrendered";

#[tokio::test]
async fn renderer_socket_returns_png() {
    if let Ok(socket) = std::env::var("CYBEREDGE_TEST_RENDERER_SOCKET") {
        let png = SocketScreenshotProbe::new(socket)
            .capture(b"<h1>CyberEdge</h1>")
            .await
            .expect("external renderer should return a PNG");
        assert!(png.starts_with(b"\x89PNG\r\n\x1a\n"));
        return;
    }

    let socket = temporary_socket();
    let listener = UnixListener::bind(&socket).expect("bind test renderer socket");
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.expect("accept renderer client");
        let length = stream.read_u32().await.expect("read request length") as usize;
        let mut html = vec![0; length];
        stream.read_exact(&mut html).await.expect("read HTML");
        assert_eq!(html, b"<h1>CyberEdge</h1>");
        stream.write_u8(0).await.expect("write status");
        stream
            .write_u32(PNG.len() as u32)
            .await
            .expect("write response length");
        stream.write_all(PNG).await.expect("write PNG");
    });

    let png = SocketScreenshotProbe::new(&socket)
        .capture(b"<h1>CyberEdge</h1>")
        .await
        .expect("renderer should return a PNG");
    assert_eq!(png, PNG);
    server.await.expect("renderer task");
    std::fs::remove_file(socket).expect("remove test socket");
}

fn temporary_socket() -> PathBuf {
    std::env::temp_dir().join(format!("cyberedge-renderer-{}.sock", Uuid::now_v7()))
}
