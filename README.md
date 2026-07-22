# CyberEdge

CyberEdge is an AI-native external attack surface engine. AI agents operate it through Skills and a typed RPC contract; humans observe through an optional read-only Web interface.

The new product is an external attack surface management platform built around a single auditable flow: discover assets, record observations, verify findings, retain evidence, and drive remediation.

## Stack

- Rust 1.97 modular monolith with Tokio and tonic
- gRPC + Protobuf as the only control-plane contract
- PostgreSQL as system of record and initial job queue
- Optional read-only enterprise Web projection
- Scanner processes isolated behind typed adapters

See [the architecture baseline](docs/architecture.md) for the design boundaries.

## Development

```bash
cargo test
cargo run
```

The local RPC server listens on `unix:///tmp/cyberedge.sock` by default. Set `CYBEREDGE_RPC_SOCKET` to use another socket path.

The current RPC baseline implements health, authorized Scope creation, Task creation, Task event streaming, cancellation, typed errors, and idempotent mutations. Persistence, capability verification, scanner execution, and the Web read model remain intentionally unimplemented.
