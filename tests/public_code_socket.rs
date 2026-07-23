use std::path::PathBuf;

use cyberedge::{PublicCodeProbe, SocketPublicCodeProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::UnixListener,
};
use uuid::Uuid;

#[tokio::test]
async fn public_code_socket_exchanges_bounded_machine_payloads() {
    let socket = temporary_socket();
    let listener = UnixListener::bind(&socket).expect("bind public code socket");
    let expected = br#"[{"query_domain":"example.com","repository":"example/repo","path":"config.yaml","name":"config.yaml","blob_sha":"0123456789abcdef0123456789abcdef01234567","html_url":"https://github.com/example/repo/blob/main/config.yaml"}]"#.to_vec();
    let response = expected.clone();
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.expect("accept public code client");
        let length = stream.read_u32().await.expect("read request length") as usize;
        let mut request = vec![0; length];
        stream.read_exact(&mut request).await.expect("read request");
        let request: serde_json::Value = serde_json::from_slice(&request).expect("request JSON");
        assert_eq!(request["domains"][0], "example.com");
        stream.write_u8(0).await.expect("write status");
        stream
            .write_u32(response.len() as u32)
            .await
            .expect("write response length");
        stream.write_all(&response).await.expect("write response");
    });

    let output = SocketPublicCodeProbe::new(&socket)
        .search(&["example.com".to_owned()])
        .await
        .expect("public code adapter response");
    assert_eq!(output, expected);
    server.await.expect("public code server task");
    std::fs::remove_file(socket).expect("remove public code test socket");
}

fn temporary_socket() -> PathBuf {
    std::env::temp_dir().join(format!("cyberedge-public-code-{}.sock", Uuid::now_v7()))
}
