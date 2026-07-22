---
name: cyberedge-observe-websites
description: Collect bounded HTTP metadata and response evidence from Web services in an explicitly authorized CyberEdge scope. Use for evidence-backed website inventory and basic HTTP fingerprint hints.
---

# CyberEdge Observe Websites

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing Scope with a non-empty authorization reference.
- Require this exact Skill/version grant with `scan.active` and `website.read`.
- Do not supply ports, arbitrary URLs, headers or redirect destinations. The server derives endpoints from authorized Scope targets and its fixed baseline.

## Workflow

1. Start `policy_service_baseline` for the authorized Scope.
2. Wait for the Task to become terminal.
3. Call `search_websites` for the same Scope.
4. Correlate Website records with `http.response` Observations and TaskReport Evidence.
5. Describe `server` as a header-derived hint, never as verified product identity.

The server probes only open `80`, `443`, `8080`, and `8443`, never follows redirects, enforces a five-second request timeout and rejects response bodies over 1 MiB. `http.error` means collection failed; it does not prove that no website exists.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-observe-websites","skill_version":"0.1.0","action":"search_websites","scope_id":"scope-id"}
```
