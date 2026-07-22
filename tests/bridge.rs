use std::sync::Arc;

use cyberedge::{Authorizer, CyberEdgeService, MemoryRepository, proto::InvocationContext};
use tokio::{io::AsyncWriteExt, net::UnixListener, process::Command};
use tokio_stream::wrappers::UnixListenerStream;
use tonic::transport::Server;

struct Allow;

impl Authorizer for Allow {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

#[tokio::test]
async fn machine_bridge_accepts_json_and_returns_json() {
    let socket =
        std::env::temp_dir().join(format!("cyberedge-bridge-{}.sock", uuid::Uuid::now_v7()));
    let listener = UnixListener::bind(&socket).unwrap();
    let server = tokio::spawn(async move {
        Server::builder()
            .add_service(
                CyberEdgeService::new(Arc::new(MemoryRepository::default()), Arc::new(Allow))
                    .server(),
            )
            .serve_with_incoming(UnixListenerStream::new(listener))
            .await
            .unwrap();
    });

    let mut child = Command::new(env!("CARGO_BIN_EXE_cyberedge-agent"))
        .env("CYBEREDGE_RPC_SOCKET", &socket)
        .stdin(std::process::Stdio::piped())
        .stdout(std::process::Stdio::piped())
        .spawn()
        .unwrap();
    child
        .stdin
        .take()
        .unwrap()
        .write_all(
            br#"{
        "request_id":"req_bridge","idempotency_key":"idem_bridge",
        "agent_id":"agent_bridge","skill_name":"cyberedge-discover-assets",
        "skill_version":"0.1.0","action":"create_scope","name":"Bridge",
        "authorization_ref":"authorization:test",
        "targets":[{"kind":"domain","value":"Example.COM"}]
    }"#,
        )
        .await
        .unwrap();
    let output = child.wait_with_output().await.unwrap();
    server.abort();
    let _ = std::fs::remove_file(socket);

    assert!(
        output.status.success(),
        "{}",
        String::from_utf8_lossy(&output.stdout)
    );
    let value: serde_json::Value = serde_json::from_slice(&output.stdout).unwrap();
    assert_eq!(value["ok"], true);
    assert_eq!(value["result"]["targets"][0]["value"], "example.com");
}
