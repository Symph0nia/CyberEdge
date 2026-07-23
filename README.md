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

The implemented core includes PostgreSQL persistence, capability-gated Scope and Task RPCs, durable Tasks/Schedules/Monitors, and a one-call Assessment workflow inspired by ARL-Next's useful pipeline semantics: readiness check, passive discovery, bounded subdomain expansion, active inventory, TLS/HTTP/Website observation, vulnerability checks, public-code intelligence, exact CPE-backed CVE correlation, registration intelligence, and explicit per-stage coverage. A completed Task never hides an unavailable adapter. The system also includes evidence-backed Findings, change monitoring, reliable webhook delivery, deterministic reports, audit records, query-only and execution Skills, a JSON machine bridge, local UDS and remote mTLS transports, and an optional OIDC-protected strictly read-only Web projection.

For a local self-hosted deployment:

```bash
docker compose up --build -d
```

## Contributing and license

Read [CONTRIBUTING.md](CONTRIBUTING.md) before proposing changes and report vulnerabilities through the process in [SECURITY.md](SECURITY.md). CyberEdge is licensed under the [Apache License 2.0](LICENSE).
