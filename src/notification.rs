use std::sync::Arc;

use async_trait::async_trait;
use serde_json::json;

use crate::repository::{OutboxEvent, PostgresRepository, RepositoryError};

#[async_trait]
pub trait NotificationSink: Send + Sync {
    async fn deliver(&self, event: &OutboxEvent) -> Result<(), String>;
}

pub struct WebhookSink {
    client: reqwest::Client,
    url: reqwest::Url,
    bearer_token: Option<String>,
}

impl WebhookSink {
    pub fn new(url: &str, bearer_token: Option<String>) -> Result<Self, String> {
        let _ = rustls::crypto::ring::default_provider().install_default();
        let url = reqwest::Url::parse(url).map_err(|error| error.to_string())?;
        if !matches!(url.scheme(), "http" | "https") {
            return Err("webhook URL must use http or https".to_owned());
        }
        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .redirect(reqwest::redirect::Policy::none())
            .user_agent("CyberEdge/0.1 notification-dispatcher")
            .build()
            .map_err(|error| error.to_string())?;
        Ok(Self {
            client,
            url,
            bearer_token,
        })
    }
}

#[async_trait]
impl NotificationSink for WebhookSink {
    async fn deliver(&self, event: &OutboxEvent) -> Result<(), String> {
        let mut request = self.client.post(self.url.clone()).json(&json!({
            "id": event.id,
            "aggregate_kind": event.aggregate_kind,
            "aggregate_id": event.aggregate_id,
            "sequence": event.sequence,
            "event_type": event.event_type,
            "payload": event.payload,
        }));
        if let Some(token) = &self.bearer_token {
            request = request.bearer_auth(token);
        }
        request
            .send()
            .await
            .map_err(|error| error.to_string())?
            .error_for_status()
            .map_err(|error| error.to_string())?;
        Ok(())
    }
}

pub struct NotificationDispatcher {
    repository: Arc<PostgresRepository>,
    sink: Arc<dyn NotificationSink>,
}

impl NotificationDispatcher {
    pub fn new(repository: Arc<PostgresRepository>, sink: Arc<dyn NotificationSink>) -> Self {
        Self { repository, sink }
    }

    pub async fn dispatch_once(&self) -> Result<bool, RepositoryError> {
        let Some(event) = self.repository.claim_outbox().await? else {
            return Ok(false);
        };
        match self.sink.deliver(&event).await {
            Ok(()) => self.repository.publish_outbox(&event.id).await?,
            Err(error) => {
                self.repository
                    .retry_outbox(&event.id, event.attempts, &error)
                    .await?
            }
        }
        Ok(true)
    }
}
