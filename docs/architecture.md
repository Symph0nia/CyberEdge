# Architecture

CyberEdge starts again as a modular monolith. One deployable API owns the domain model and one worker pool executes scan jobs. PostgreSQL is the source of truth and the initial job queue.

## Baseline

- Rust 1.97: API, scheduler, workers, and scanner adapters
- Axum + Tokio: HTTP and asynchronous runtime
- SQLx: PostgreSQL access once the first persistent domain model lands
- PostgreSQL: assets, observations, findings, jobs, and audit events
- React 19.2 + TypeScript + Vite 8.1: operator console
- OpenAPI: the contract between API and console
- OCI images + Docker Compose: the first deployment target

## Domain flow

`scope -> asset -> observation -> finding -> evidence -> remediation`

Jobs are claimed with PostgreSQL row locking. A dedicated broker is introduced only when measured throughput proves PostgreSQL insufficient.

Scanner tools run behind adapters with explicit timeouts, resource limits, and normalized output. Raw tool output is evidence, never the domain model.

## Boundaries

- `src/`: API, domain, and infrastructure code
- `web/`: operator console
- `docs/`: architecture decisions and operator documentation

No microservices, plugin framework, or distributed workflow engine until a real constraint requires one.
