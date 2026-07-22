pub mod proto {
    tonic::include_proto!("cyberedge.v1");
}

mod service;

pub use service::CyberEdgeService;
