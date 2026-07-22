pub mod proto {
    tonic::include_proto!("cyberedge.v1");
}

mod policy;
mod repository;
mod service;
mod web;
mod worker;

pub use policy::{Authorizer, PolicyError, StaticAuthorizer};
pub use repository::{MemoryRepository, PostgresRepository, Repository, RepositoryError};
pub use service::CyberEdgeService;
pub use web::{read_only_router, serve_read_only_web};
pub use worker::{
    BASELINE_SERVICE_PORTS, CertificateProbe, CertificateSource, CrtShSource, DiscoveryWorker,
    DnsResolver, PortConnector, SystemCertificateProbe, SystemDnsResolver, SystemPortConnector,
};
