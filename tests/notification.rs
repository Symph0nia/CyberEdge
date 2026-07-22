use std::sync::Arc;

use async_trait::async_trait;
use cyberedge::{
    NotificationDispatcher, NotificationSink, OutboxEvent, PostgresRepository, WebhookSink,
};
use sqlx::PgPool;
use tokio::io::{AsyncReadExt, AsyncWriteExt};

struct FailingSink;

#[async_trait]
impl NotificationSink for FailingSink {
    async fn deliver(&self, _event: &OutboxEvent) -> Result<(), String> {
        Err("receiver unavailable".to_owned())
    }
}

#[tokio::test]
async fn delivers_and_retries_outbox_events() {
    let Ok(database_url) = std::env::var("TEST_DATABASE_URL") else {
        return;
    };
    let repository = Arc::new(PostgresRepository::connect(&database_url).await.unwrap());
    let pool = PgPool::connect(&database_url).await.unwrap();
    sqlx::query("TRUNCATE outbox_events")
        .execute(&pool)
        .await
        .unwrap();
    insert_event(&pool, "event_delivery").await;

    let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
    let address = listener.local_addr().unwrap();
    let receiver = tokio::spawn(async move {
        let (mut stream, _) = listener.accept().await.unwrap();
        let mut request = Vec::new();
        let mut buffer = [0_u8; 4096];
        loop {
            let read = stream.read(&mut buffer).await.unwrap();
            request.extend_from_slice(&buffer[..read]);
            let complete = request
                .windows(4)
                .position(|value| value == b"\r\n\r\n")
                .is_some_and(|header_end| {
                    let headers = String::from_utf8_lossy(&request[..header_end]);
                    let content_length = headers
                        .lines()
                        .find_map(|line| {
                            let (name, value) = line.split_once(':')?;
                            name.eq_ignore_ascii_case("content-length")
                                .then(|| value.trim().parse::<usize>().ok())
                                .flatten()
                        })
                        .unwrap_or_default();
                    request.len() >= header_end + 4 + content_length
                });
            if read == 0 || complete {
                break;
            }
        }
        let request = String::from_utf8_lossy(&request);
        assert!(
            request
                .to_ascii_lowercase()
                .contains("authorization: bearer local-secret")
        );
        assert!(request.contains("event_delivery"));
        stream
            .write_all(b"HTTP/1.1 204 No Content\r\nConnection: close\r\n\r\n")
            .await
            .unwrap();
    });
    let sink = Arc::new(
        WebhookSink::new(
            &format!("http://{address}/events"),
            Some("local-secret".to_owned()),
        )
        .unwrap(),
    );
    let dispatcher = NotificationDispatcher::new(repository.clone(), sink);
    assert!(dispatcher.dispatch_once().await.unwrap());
    receiver.await.unwrap();
    let delivered: bool = sqlx::query_scalar(
        "SELECT published_at IS NOT NULL AND attempts = 1 FROM outbox_events WHERE id = $1",
    )
    .bind("event_delivery")
    .fetch_one(&pool)
    .await
    .unwrap();
    assert!(delivered);

    insert_event(&pool, "event_retry").await;
    let dispatcher = NotificationDispatcher::new(repository, Arc::new(FailingSink));
    assert!(dispatcher.dispatch_once().await.unwrap());
    let retry: (i32, String, bool) = sqlx::query_as(
        "SELECT attempts, last_error, next_attempt_at > now()
         FROM outbox_events WHERE id = $1",
    )
    .bind("event_retry")
    .fetch_one(&pool)
    .await
    .unwrap();
    assert_eq!(retry.0, 1);
    assert_eq!(retry.1, "receiver unavailable");
    assert!(retry.2);
}

async fn insert_event(pool: &PgPool, id: &str) {
    sqlx::query(
        "INSERT INTO outbox_events
         (id, aggregate_kind, aggregate_id, event_type, payload)
         VALUES ($1, 'test', 'aggregate', 'test.event', '{\"ok\":true}')",
    )
    .bind(id)
    .execute(pool)
    .await
    .unwrap();
}
