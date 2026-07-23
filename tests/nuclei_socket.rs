use std::path::PathBuf;

use cyberedge::{NucleiProbe, SocketNucleiProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::UnixListener,
};
use uuid::Uuid;

#[tokio::test]
async fn nuclei_socket_exchanges_bounded_machine_payloads() {
    if let Ok(socket) = std::env::var("CYBEREDGE_TEST_NUCLEI_SOCKET") {
        let target =
            std::env::var("CYBEREDGE_TEST_NUCLEI_TARGET").expect("external nuclei test target");
        let output = SocketNucleiProbe::new(socket)
            .scan(&[target])
            .await
            .expect("external nuclei adapter response");
        assert!(String::from_utf8_lossy(&output).contains("default-nginx-page"));
        let result: serde_json::Value = serde_json::from_slice(
            output
                .split(|byte| *byte == b'\n')
                .find(|line| !line.is_empty())
                .expect("external nuclei result"),
        )
        .expect("external nuclei JSONL");
        assert_eq!(result["type"], "http");
        assert_eq!(result["info"]["severity"], "info");
        for forbidden in ["request", "response", "template-encoded", "interaction"] {
            assert!(
                result.get(forbidden).is_none(),
                "forbidden field returned: {forbidden}"
            );
        }
        return;
    }
    let socket = temporary_socket();
    let listener = UnixListener::bind(&socket).expect("bind test nuclei socket");
    let expected = br#"{"template-id":"test","info":{"severity":"high"}}"#.to_vec();
    let response = expected.clone();
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.expect("accept nuclei client");
        let length = stream.read_u32().await.expect("read request length") as usize;
        let mut request = vec![0; length];
        stream.read_exact(&mut request).await.expect("read request");
        let request: serde_json::Value = serde_json::from_slice(&request).expect("request JSON");
        assert_eq!(request["targets"][0], "https://example.com/");
        stream.write_u8(0).await.expect("write status");
        stream
            .write_u32(response.len() as u32)
            .await
            .expect("write response length");
        stream.write_all(&response).await.expect("write response");
    });

    let output = SocketNucleiProbe::new(&socket)
        .scan(&["https://example.com/".to_owned()])
        .await
        .expect("nuclei adapter response");
    assert_eq!(output, expected);
    server.await.expect("nuclei server task");
    std::fs::remove_file(socket).expect("remove nuclei test socket");
}

fn temporary_socket() -> PathBuf {
    std::env::temp_dir().join(format!("cyberedge-nuclei-{}.sock", Uuid::now_v7()))
}
