---
name: cyberedge-discover-services
description: Run controlled TCP connect service discovery through the CyberEdge Agent Protocol for an explicitly authorized Scope. Use when an AI agent must identify exposed baseline services on approved domain or IP targets; never use for ambiguous targets, vulnerability exploitation, arbitrary port ranges, or targets without a human-provided authorization reference.
---

# Discover CyberEdge Services

Use `cyberedge-agent` as the JSON machine bridge. This Skill performs active network connections and therefore requires both an existing authorized Scope and the `scan.active` capability.

## Preconditions

- Confirm the exact Scope and its human-provided authorization reference before starting.
- Refuse ambiguous ownership, third-party targets, arbitrary port lists, or requests to evade controls.
- Use a fresh `request_id` and `idempotency_key` for each mutation.
- Send this Skill's declared name and version in every invocation context.

## Workflow

1. Call `GetScope` and verify every target remains inside the requested authorization boundary.
2. Call `StartScan` with `policy_service_baseline`. The server controls the baseline TCP port set; do not attempt to override it.
3. Call `WatchTask` and continue from the last accepted sequence after reconnecting.
4. Stop at `completed`, `failed`, or `canceled`.
5. Call `SearchServices` with the Scope ID and `GetTaskReport` with the completed Task ID.
6. Report service hints as port-based hints, not banner-verified product identities.
7. Preserve coverage errors and evidence references. Never interpret a timeout as proof that a service does not exist.

Example:

```json
{"request_id":"req_...","idempotency_key":"idem_...","agent_id":"agent_...","skill_name":"cyberedge-discover-services","skill_version":"0.1.0","action":"start_scan","scope_id":"scope_...","policy_id":"policy_service_baseline"}
```

## Failure handling

- On `CAPABILITY_DENIED`, stop; do not switch identities or Skills.
- On authorization or Scope validation failure, stop and request a corrected authorization boundary.
- Retry only typed retryable failures.
- Do not fall back to external scanners or direct socket commands.

## Output

Return the Scope ID, Task ID, terminal state, open TCP ports, service hints, evidence references, and coverage limitations.
