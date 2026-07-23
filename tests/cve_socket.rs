use std::path::PathBuf;

use cyberedge::{CveProbe, SocketCveProbe};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::UnixListener,
};
use uuid::Uuid;

#[tokio::test]
async fn cve_socket_exchanges_bounded_exact_cpe_payloads() {
    if let Ok(socket) = std::env::var("CYBEREDGE_TEST_CVE_SOCKET") {
        let cpe = "cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*".to_owned();
        let output = SocketCveProbe::new(socket)
            .query(std::slice::from_ref(&cpe))
            .await
            .expect("external CVE adapter response");
        let results: serde_json::Value = serde_json::from_slice(&output).expect("CVE JSON");
        let results = results.as_array().expect("CVE result array");
        assert!(!results.is_empty());
        assert!(results.iter().all(|result| result["cpe_name"] == cpe));
        assert!(results.iter().all(|result| {
            result["cve_id"]
                .as_str()
                .is_some_and(|value| value.starts_with("CVE-"))
        }));
        assert!(
            results
                .iter()
                .all(|result| result.get("configurations").is_none())
        );
        return;
    }
    let socket = temporary_socket();
    let listener = UnixListener::bind(&socket).expect("bind CVE socket");
    let expected = br#"[{"cpe_name":"cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*","cve_id":"CVE-2099-1234"}]"#.to_vec();
    let response = expected.clone();
    let server = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.expect("accept CVE client");
        let length = stream.read_u32().await.expect("read request length") as usize;
        let mut request = vec![0; length];
        stream.read_exact(&mut request).await.expect("read request");
        let request: serde_json::Value = serde_json::from_slice(&request).expect("request JSON");
        assert_eq!(
            request["cpe_names"][0],
            "cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*"
        );
        stream.write_u8(0).await.expect("write status");
        stream
            .write_u32(response.len() as u32)
            .await
            .expect("write response length");
        stream.write_all(&response).await.expect("write response");
    });

    let output = SocketCveProbe::new(&socket)
        .query(&["cpe:2.3:a:wordpress:wordpress:6.8:*:*:*:*:*:*:*".to_owned()])
        .await
        .expect("CVE adapter response");
    assert_eq!(output, expected);
    server.await.expect("CVE server task");
    std::fs::remove_file(socket).expect("remove CVE test socket");
}

fn temporary_socket() -> PathBuf {
    std::env::temp_dir().join(format!("cyberedge-cve-{}.sock", Uuid::now_v7()))
}
