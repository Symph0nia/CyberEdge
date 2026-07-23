use std::{path::PathBuf, sync::Arc};

use axum::{
    body::Body,
    http::{Request, StatusCode},
};
use base64::{Engine, engine::general_purpose::URL_SAFE_NO_PAD};
use cyberedge::{MemoryRepository, OidcAccess, WebAccess, read_only_router};
use jsonwebtoken::{Algorithm, EncodingKey, Header, encode, jwk::JwkSet};
use rsa::{RsaPrivateKey, pkcs1::EncodeRsaPrivateKey, traits::PublicKeyParts};
use serde_json::json;
use tower::ServiceExt;

#[tokio::test]
async fn web_exposes_only_read_routes_with_security_headers() {
    let app = read_only_router(
        Arc::new(MemoryRepository::default()),
        PathBuf::from("web/dist"),
        WebAccess::InsecureLocal,
    );
    let response = app
        .clone()
        .oneshot(
            Request::get("/api/v1/overview")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    assert_eq!(response.status(), StatusCode::OK);
    assert!(response.headers().contains_key("content-security-policy"));
    assert_eq!(response.headers()["x-content-type-options"], "nosniff");

    let mutation = app
        .clone()
        .oneshot(
            Request::post("/api/v1/overview")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    assert_eq!(mutation.status(), StatusCode::METHOD_NOT_ALLOWED);

    let unknown_api = app
        .oneshot(Request::get("/api/v1/mutate").body(Body::empty()).unwrap())
        .await
        .unwrap();
    assert_eq!(unknown_api.status(), StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn web_requires_valid_oidc_roles_and_separates_evidence_access() {
    let private = RsaPrivateKey::new(&mut rand::thread_rng(), 2048).unwrap();
    let jwks: JwkSet = serde_json::from_value(json!({"keys": [{
        "kty": "RSA",
        "alg": "RS256",
        "use": "sig",
        "kid": "test-key",
        "n": URL_SAFE_NO_PAD.encode(private.n().to_bytes_be()),
        "e": URL_SAFE_NO_PAD.encode(private.e().to_bytes_be())
    }]}))
    .unwrap();
    let access = WebAccess::Oidc(Arc::new(OidcAccess::new(
        "https://identity.example".to_owned(),
        "cyberedge-web".to_owned(),
        "roles".to_owned(),
        "cyberedge.read".to_owned(),
        "cyberedge.evidence.read".to_owned(),
        jwks,
    )));
    let app = read_only_router(
        Arc::new(MemoryRepository::default()),
        PathBuf::from("web/dist"),
        access,
    );
    let key = EncodingKey::from_rsa_der(private.to_pkcs1_der().unwrap().as_bytes());
    let mut header = Header::new(Algorithm::RS256);
    header.kid = Some("test-key".to_owned());
    let now = jsonwebtoken::get_current_timestamp();
    let token = encode(
        &header,
        &json!({
            "iss": "https://identity.example",
            "aud": "cyberedge-web",
            "sub": "viewer@example.com",
            "exp": now + 300,
            "roles": ["cyberedge.read"]
        }),
        &key,
    )
    .unwrap();

    let anonymous = app
        .clone()
        .oneshot(
            Request::get("/api/v1/overview")
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    assert_eq!(anonymous.status(), StatusCode::UNAUTHORIZED);

    let overview = app
        .clone()
        .oneshot(
            Request::get("/api/v1/overview")
                .header("authorization", format!("Bearer {token}"))
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    assert_eq!(overview.status(), StatusCode::OK);

    let evidence = app
        .oneshot(
            Request::get("/api/v1/evidence/example")
                .header("authorization", format!("Bearer {token}"))
                .body(Body::empty())
                .unwrap(),
        )
        .await
        .unwrap();
    assert_eq!(evidence.status(), StatusCode::FORBIDDEN);
}
