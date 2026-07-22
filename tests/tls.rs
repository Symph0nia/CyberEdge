use std::sync::Arc;

use cyberedge::{
    Authorizer, CyberEdgeService, MemoryRepository,
    proto::{InvocationContext, cyber_edge_client::CyberEdgeClient},
};
use rcgen::{
    BasicConstraints, CertificateParams, CertifiedIssuer, ExtendedKeyUsagePurpose, IsCa, KeyPair,
    KeyUsagePurpose,
};
use tokio::net::TcpListener;
use tokio_stream::wrappers::TcpListenerStream;
use tonic::transport::{Certificate, ClientTlsConfig, Endpoint, Identity, Server, ServerTlsConfig};

struct Allow;

impl Authorizer for Allow {
    fn authorize(&self, _context: &InvocationContext, _capability: &str) -> bool {
        true
    }
}

#[tokio::test]
async fn remote_transport_requires_a_valid_client_certificate() {
    let certificates = certificates();
    let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
    let address = listener.local_addr().unwrap();
    let server_identity = Identity::from_pem(&certificates.server_cert, &certificates.server_key);
    let client_ca = Certificate::from_pem(&certificates.ca_cert);
    let server = tokio::spawn(async move {
        Server::builder()
            .tls_config(
                ServerTlsConfig::new()
                    .identity(server_identity)
                    .client_ca_root(client_ca),
            )
            .unwrap()
            .add_service(
                CyberEdgeService::new(Arc::new(MemoryRepository::default()), Arc::new(Allow))
                    .server(),
            )
            .serve_with_incoming(TcpListenerStream::new(listener))
            .await
            .unwrap();
    });
    let endpoint = format!("https://localhost:{}", address.port());
    let ca = Certificate::from_pem(&certificates.ca_cert);

    let anonymous = Endpoint::from_shared(endpoint.clone())
        .unwrap()
        .tls_config(
            ClientTlsConfig::new()
                .domain_name("localhost")
                .ca_certificate(ca.clone()),
        )
        .unwrap()
        .connect()
        .await
        .unwrap();
    assert!(CyberEdgeClient::new(anonymous).health(()).await.is_err());

    let authenticated = Endpoint::from_shared(endpoint)
        .unwrap()
        .tls_config(
            ClientTlsConfig::new()
                .domain_name("localhost")
                .ca_certificate(ca)
                .identity(Identity::from_pem(
                    &certificates.client_cert,
                    &certificates.client_key,
                )),
        )
        .unwrap()
        .connect()
        .await
        .unwrap();
    assert_eq!(
        CyberEdgeClient::new(authenticated)
            .health(())
            .await
            .unwrap()
            .into_inner()
            .status,
        "ok"
    );
    server.abort();
}

struct Certificates {
    ca_cert: String,
    server_cert: String,
    server_key: String,
    client_cert: String,
    client_key: String,
}

fn certificates() -> Certificates {
    let ca_key = KeyPair::generate().unwrap();
    let mut ca_params = CertificateParams::new(vec!["CyberEdge test CA".to_owned()]).unwrap();
    ca_params.is_ca = IsCa::Ca(BasicConstraints::Unconstrained);
    ca_params.key_usages = vec![
        KeyUsagePurpose::KeyCertSign,
        KeyUsagePurpose::DigitalSignature,
    ];
    let ca = CertifiedIssuer::self_signed(ca_params, ca_key).unwrap();

    let server_key = KeyPair::generate().unwrap();
    let mut server_params = CertificateParams::new(vec!["localhost".to_owned()]).unwrap();
    server_params.extended_key_usages = vec![ExtendedKeyUsagePurpose::ServerAuth];
    let server_cert = server_params.signed_by(&server_key, &ca).unwrap();

    let client_key = KeyPair::generate().unwrap();
    let mut client_params = CertificateParams::new(vec!["cyberedge-agent".to_owned()]).unwrap();
    client_params.extended_key_usages = vec![ExtendedKeyUsagePurpose::ClientAuth];
    let client_cert = client_params.signed_by(&client_key, &ca).unwrap();

    Certificates {
        ca_cert: ca.pem(),
        server_cert: server_cert.pem(),
        server_key: server_key.serialize_pem(),
        client_cert: client_cert.pem(),
        client_key: client_key.serialize_pem(),
    }
}
