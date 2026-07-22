# CyberEdge

CyberEdge is restarting from a clean history and a deliberately small architecture.

The new product is an external attack surface management platform built around a single auditable flow: discover assets, record observations, verify findings, retain evidence, and drive remediation.

## Stack

- Rust 1.97 modular monolith with Axum and Tokio
- PostgreSQL as system of record and initial job queue
- React 19.2, TypeScript, and Vite 8.1
- Scanner processes isolated behind typed adapters

See [the architecture baseline](docs/architecture.md) for the design boundaries.

## Development

```bash
cargo test
cargo run
```

```bash
cd web
npm install
npm run dev
```

The API health endpoint is `GET /api/v1/health`.

> This repository is at the architecture-reset baseline. Product features will be introduced through new issues against this codebase.
