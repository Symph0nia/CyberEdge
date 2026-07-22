use cyberedge::{
    Authorizer, CyberEdgeService, MemoryRepository,
    proto::{
        CreateScopeRequest, InvocationContext, ScopeTarget, TargetKind,
        cyber_edge_client::CyberEdgeClient,
    },
};
use std::sync::Arc;
use tokio::net::TcpListener;
use tokio_stream::wrappers::TcpListenerStream;
use tonic::transport::Server;

struct TestAuthorizer;

impl Authorizer for TestAuthorizer {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

#[tokio::test]
async fn serves_rpc_contract() {
    let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
    let address = listener.local_addr().unwrap();
    let server = tokio::spawn(async move {
        Server::builder()
            .add_service(
                CyberEdgeService::new(
                    Arc::new(MemoryRepository::default()),
                    Arc::new(TestAuthorizer),
                )
                .server(),
            )
            .serve_with_incoming(TcpListenerStream::new(listener))
            .await
            .unwrap();
    });
    let mut client = CyberEdgeClient::connect(format!("http://{address}"))
        .await
        .unwrap();

    let health = client.health(()).await.unwrap().into_inner();
    assert_eq!(health.status, "ok");

    let scope = client
        .create_scope(CreateScopeRequest {
            context: Some(InvocationContext {
                request_id: "req_rpc_test".to_owned(),
                idempotency_key: "idem_rpc_test".to_owned(),
                agent_id: "agent_rpc_test".to_owned(),
                skill_name: "asset-discovery".to_owned(),
                skill_version: "1.0.0".to_owned(),
            }),
            name: "RPC test".to_owned(),
            authorization_ref: "authorization:rpc-test".to_owned(),
            targets: vec![ScopeTarget {
                kind: TargetKind::Domain.into(),
                value: "Example.COM".to_owned(),
            }],
        })
        .await
        .unwrap()
        .into_inner();

    assert_eq!(scope.targets[0].value, "example.com");
    server.abort();
}
