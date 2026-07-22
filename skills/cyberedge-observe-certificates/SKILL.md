---
name: cyberedge-observe-certificates
description: Collect and inspect TLS leaf certificates for an explicitly authorized CyberEdge scope. Use only when an authorization reference already exists and active service discovery is permitted.
---

# CyberEdge Observe Certificates

Use `cyberedge-agent` as the only control-plane interface. Humans do not operate this bridge directly.

## Preconditions

- Require an existing Scope with a non-empty `authorization_ref` covering every target.
- Require the exact Skill grant `cyberedge-observe-certificates` with `scan.active`.
- Never widen a Scope, accept arbitrary ports, invoke an external scanner, or bypass RPC policy.

## Workflow

1. Call `start_scan` with the authorized `scope_id` and `policy_service_baseline`.
2. Watch or read the Task until it reaches a terminal state.
3. Call `search_certificates` with the same `scope_id`.
4. Call `get_task_report` to correlate each Certificate with its `tls.certificate` Observation and DER Evidence.
5. Treat Subject, Issuer, SAN and validity fields as observations. Report the SHA-256 fingerprint when identifying a certificate.

The policy probes TLS only on open baseline HTTPS ports `443` and `8443`. A handshake or parse failure is retained as `tls.error`; it must not be presented as evidence that no certificate exists.

## Machine envelope

Send one JSON object on standard input. Required common fields are `request_id`, `idempotency_key`, `agent_id`, `skill_name`, and `skill_version`.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-observe-certificates","skill_version":"0.1.0","action":"search_certificates","scope_id":"scope-id"}
```
