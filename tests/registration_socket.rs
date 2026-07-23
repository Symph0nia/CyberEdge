use std::path::PathBuf;

use cyberedge::{RegistrationProbe, SocketRegistrationProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::UnixListener,
};
use uuid::Uuid;

#[tokio::test]
async fn registration_socket_exchanges_bounded_domain_payloads() {
    let socket = temporary_socket();
    let listener = UnixListener::bind(&socket).expect("bind registration socket");
    let expected =
        br#"{"coverage":[{"domain":"example.cn","status":"complete"}],"records":[]}"#.to_vec();
    let response = expected.clone();
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.expect("accept registration client");
        let length = stream.read_u32().await.expect("read request length") as usize;
        let mut request = vec![0; length];
        stream.read_exact(&mut request).await.expect("read request");
        let request: serde_json::Value = serde_json::from_slice(&request).expect("request JSON");
        assert_eq!(request["domains"][0], "example.cn");
        stream.write_u8(0).await.expect("write status");
        stream
            .write_u32(response.len() as u32)
            .await
            .expect("write response length");
        stream.write_all(&response).await.expect("write response");
    });

    let output = SocketRegistrationProbe::new(&socket)
        .lookup(&["example.cn".to_owned()])
        .await
        .expect("registration adapter response");
    assert_eq!(output, expected);
    server.await.expect("registration server task");
    std::fs::remove_file(socket).expect("remove registration test socket");
}

fn temporary_socket() -> PathBuf {
    std::env::temp_dir().join(format!("cyberedge-registration-{}.sock", Uuid::now_v7()))
}
