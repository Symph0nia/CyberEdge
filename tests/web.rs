use std::{path::PathBuf, sync::Arc};

use axum::{
    body::Body,
    http::{Request, StatusCode},
};
use cyberedge::{MemoryRepository, read_only_router};
use tower::ServiceExt;

#[tokio::test]
async fn web_exposes_only_read_routes_with_security_headers() {
    let app = read_only_router(
        Arc::new(MemoryRepository::default()),
        PathBuf::from("web/dist"),
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
