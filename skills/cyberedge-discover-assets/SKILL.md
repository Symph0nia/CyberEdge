---
name: cyberedge-discover-assets
description: Create an authorized CyberEdge scope, run passive asset discovery, follow task events, and retrieve evidence-backed results through the CyberEdge Agent Protocol. Use when an AI agent must inventory domains or IP space that has an explicit authorization reference; never use for active probing or targets outside that scope.
---

# Discover CyberEdge Assets

Use `cyberedge-agent` as the machine bridge to the CyberEdge gRPC contract. Send exactly one JSON envelope on stdin and parse JSON Lines from stdout. Do not pass human-oriented flags, access the database, invoke scanners, or edit the read-only Web.

## Preconditions

- Obtain the human-provided authorization reference and exact targets.
- Refuse ambiguous ownership or authorization.
- Use a fresh `request_id` and `idempotency_key` for each mutation. Reuse the same key only when retrying the identical request.
- Send the name and version declared in `manifest.json` in every invocation context.

## Workflow

1. Call `CreateScope` with normalized target kinds and the authorization reference.
2. Call `StartScan` with capability `scan.passive`. Use `policy_passive_inventory` for broad passive inventory (DNS plus Certificate Transparency), or `policy_passive_dns` when only direct DNS resolution is requested.
3. Call `WatchTask` from sequence `0`; on reconnect, continue after the last accepted sequence.
4. Stop when the task reaches `completed`, `failed`, or `canceled`.
5. Call `GetTaskReport` after `task.completed` to retrieve the deterministic report bundle.
6. Use `SearchAudit` when the requester needs invocation provenance.
7. Report only facts linked to evidence. Separate errors and coverage gaps from confirmed absence.

For recurring monitoring, call `CreateSchedule` with the existing scope and passive policy. The minimum interval is 60 seconds. A Schedule never performs discovery itself: each due occurrence creates a normal Task, so follow and report its `last_task_id` through the same task workflow. Use `SearchSchedules` to inspect recurrence state.

Use snake-case action names. Every envelope includes `request_id`, `idempotency_key`, `agent_id`, `skill_name`, and `skill_version`. Example:

```json
{"request_id":"req_...","idempotency_key":"idem_...","agent_id":"agent_...","skill_name":"cyberedge-discover-assets","skill_version":"0.1.0","action":"start_scan","scope_id":"scope_...","policy_id":"policy_passive_inventory"}
```

Set `CYBEREDGE_RPC_SOCKET` only when the service does not use `/tmp/cyberedge.sock`.

## Failure handling

- On `CAPABILITY_DENIED`, stop and report the missing capability; do not try another identity or Skill.
- On `IDEMPOTENCY_KEY_REUSED`, generate a new key only if the intended input changed.
- On scope validation failure, narrow or correct the target; never bypass validation.
- Retry only when the typed error sets `retryable=true`.
- Cancel a running task only when the requester explicitly asks or continued execution is unsafe.

## Output

Return the scope ID, task ID, terminal state, discovered asset counts by kind, evidence references, coverage limitations, and relevant audit request IDs. Never claim that an unobserved asset does not exist.
