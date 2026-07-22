---
name: cyberedge-discover-assets
description: Create an authorized CyberEdge scope, run passive asset discovery, follow task events, and retrieve evidence-backed results through the CyberEdge Agent Protocol. Use when an AI agent must inventory domains or IP space that has an explicit authorization reference; never use for active probing or targets outside that scope.
---

# Discover CyberEdge Assets

Use only the CyberEdge gRPC contract in `proto/cyberedge/v1/cyberedge.proto`. Do not access the database, invoke scanners, or edit the read-only Web.

## Preconditions

- Obtain the human-provided authorization reference and exact targets.
- Refuse ambiguous ownership or authorization.
- Use a fresh `request_id` and `idempotency_key` for each mutation. Reuse the same key only when retrying the identical request.
- Send the name and version declared in `manifest.json` in every invocation context.

## Workflow

1. Call `CreateScope` with normalized target kinds and the authorization reference.
2. Call `StartScan` with capability `scan.passive` and policy `policy_passive_dns`.
3. Call `WatchTask` from sequence `0`; on reconnect, continue after the last accepted sequence.
4. Stop when the task reaches `completed`, `failed`, or `canceled`.
5. Query assets, observations, and evidence through protocol read methods.
6. Report only facts linked to evidence. Separate errors and coverage gaps from confirmed absence.

## Failure handling

- On `CAPABILITY_DENIED`, stop and report the missing capability; do not try another identity or Skill.
- On `IDEMPOTENCY_KEY_REUSED`, generate a new key only if the intended input changed.
- On scope validation failure, narrow or correct the target; never bypass validation.
- Retry only when the typed error sets `retryable=true`.
- Cancel a running task only when the requester explicitly asks or continued execution is unsafe.

## Output

Return the scope ID, task ID, terminal state, discovered asset counts by kind, evidence references, coverage limitations, and relevant audit request IDs. Never claim that an unobserved asset does not exist.
