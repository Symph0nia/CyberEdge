pub mod proto {
    tonic::include_proto!("cyberedge.v1");
}

mod intelligence;
mod notification;
mod nuclei;
mod policy;
mod repository;
mod service;
mod web;
mod worker;

pub use intelligence::{
    CveProbe, GitHubPublicCodeProbe, NvdCveProbe, PublicCodeProbe, SocketCveProbe,
    SocketPublicCodeProbe,
};
pub use notification::{NotificationDispatcher, NotificationSink, WebhookSink};
pub use nuclei::{NucleiProbe, SocketNucleiProbe, SystemNucleiProbe};
pub use policy::{Authorizer, PolicyError, StaticAuthorizer};
pub use repository::{
    MemoryRepository, OutboxEvent, PostgresRepository, Repository, RepositoryError,
};
pub use service::CyberEdgeService;
pub use web::{read_only_router, serve_read_only_web};
pub use worker::{
    BASELINE_SERVICE_PORTS, CertificateProbe, CertificateSource, CrtShSource, DiscoveryWorker,
    DnsResolver, PortConnector, ScreenshotProbe, SocketScreenshotProbe, SystemCertificateProbe,
    SystemDnsResolver, SystemPortConnector, SystemScreenshotProbe, SystemWebsiteProbe, WebSnapshot,
    WebsiteProbe,
};
