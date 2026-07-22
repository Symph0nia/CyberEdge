# CyberEdge

CyberEdge is an AI-native external attack surface engine. AI agents operate it through Skills and a typed RPC contract; humans observe through an optional read-only Web interface.

The new product is an external attack surface management platform built around a single auditable flow: discover assets, record observations, verify findings, retain evidence, and drive remediation.

## Stack

- Rust 1.97 modular monolith with Tokio and tonic
- gRPC + Protobuf as the only control-plane contract
- PostgreSQL as system of record and initial job queue
- Optional read-only enterprise Web projection
- Scanner processes isolated behind typed adapters

See [the accepted product architecture](docs/ai-native-architecture.md), [implementation architecture](docs/architecture.md), and [operations guide](docs/operations.md).

## Development

```bash
cargo test
cd web && npm run build
DATABASE_URL=postgres://... \
CYBEREDGE_AGENT_POLICY=config/agents.example.toml \
cargo run
```

The local RPC server listens on `unix:///tmp/cyberedge.sock` by default. Set `CYBEREDGE_RPC_SOCKET` to use another socket path.

The first vertical slice implements PostgreSQL persistence and migrations, capability-gated Scope and Task RPCs, durable Task execution, passive DNS discovery, Asset/Observation/Evidence retention, audit/outbox records, typed errors, idempotent mutations, a JSON machine bridge, local UDS and remote mTLS transports, and an optional strictly read-only Web projection.

For a local self-hosted deployment:

```bash
docker compose up --build -d
```
