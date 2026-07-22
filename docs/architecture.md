# Architecture

CyberEdge is a single-organization, self-hosted modular monolith. AI Agents are the only operators. Skills call the Rust Core through gRPC; humans use an optional strictly read-only Web projection.

## Baseline

- Rust 1.97: API, scheduler, workers, and scanner adapters
- tonic + Tokio: gRPC and asynchronous runtime
- Protobuf: the only control-plane contract
- SQLx: transactional PostgreSQL repositories and migrations
- PostgreSQL: assets, observations, findings, tasks, and audit events
- Read Model: isolated query projection for the optional enterprise Web
- OCI images + Docker Compose: the first deployment target

## Domain flow

`scope -> asset -> observation -> finding -> evidence -> remediation`

Tasks are claimed with `FOR UPDATE SKIP LOCKED`. State changes, task events, discovery records, and outbox events are committed transactionally. A dedicated broker is introduced only when measured throughput proves PostgreSQL insufficient.

Scanner tools run behind adapters with explicit timeouts, resource limits, and normalized output. Raw tool output is evidence, never the domain model.

The initial passive adapter uses the system DNS resolver. It records domain answers, discovered IP assets, lookup failures, and empty answers as immutable JSON evidence. Active probing is not implemented.

Local Agent calls use a Unix Domain Socket. Remote calls use HTTP/2 with mandatory mutual TLS. `cyberedge-agent` is a JSON stdin/stdout bridge intended for Skills, not a human command interface.

The optional Web reads a bounded projection from the repository. Its HTTP router registers only GET endpoints and static files; all state changes remain exclusive to gRPC. It binds only when explicitly enabled.

## Boundaries

- `proto/`: versioned Agent RPC contract
- `src/`: domain, RPC, task engine, and infrastructure code
- `web/`: optional read-only observation interface
- `docs/`: architecture decisions and operator documentation

No human CLI, mutable Web console, multi-tenancy, microservices, message broker, plugin framework, or distributed workflow engine until a real constraint requires one.
