# Architecture

CyberEdge is a single-organization, self-hosted modular monolith. AI Agents are the only operators. Skills call the Rust Core through gRPC; humans use an optional strictly read-only Web projection.

## Baseline

- Rust 1.97: API, scheduler, workers, and scanner adapters
- tonic + Tokio: gRPC and asynchronous runtime
- Protobuf: the only control-plane contract
- SQLx: PostgreSQL access once the first persistent domain model lands
- PostgreSQL: assets, observations, findings, tasks, and audit events
- Read Model: isolated query projection for the optional enterprise Web
- OCI images + Docker Compose: the first deployment target

## Domain flow

`scope -> asset -> observation -> finding -> evidence -> remediation`

Tasks are claimed with PostgreSQL row locking. A dedicated broker is introduced only when measured throughput proves PostgreSQL insufficient.

Scanner tools run behind adapters with explicit timeouts, resource limits, and normalized output. Raw tool output is evidence, never the domain model.

## Boundaries

- `proto/`: versioned Agent RPC contract
- `src/`: domain, RPC, task engine, and infrastructure code
- `web/`: optional read-only observation interface
- `docs/`: architecture decisions and operator documentation

No human CLI, mutable Web console, multi-tenancy, microservices, message broker, plugin framework, or distributed workflow engine until a real constraint requires one.
