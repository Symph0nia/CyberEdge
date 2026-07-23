---
name: cyberedge-inspect-results
description: Inspect CyberEdge task state, deterministic reports, inventory, findings, evidence, monitor changes, and audit provenance without performing mutations. Use when an AI agent must explain what an existing CyberEdge task observed, verify a claim against retained Evidence, diagnose coverage gaps, compare scheduled exposure changes, or trace which Agent and Skill produced a result.
---

# Inspect CyberEdge Results

Use `cyberedge-agent` as the machine bridge. Send exactly one JSON envelope on stdin and parse JSON Lines from stdout. This Skill is query-only: never create, cancel, retry, schedule, report, or otherwise mutate a resource.

## Preconditions

- Obtain at least one stable identifier: `task_id`, `scope_id`, `schedule_id`, or `evidence_id`.
- Read the Skill version from `manifest.json` and send it in every invocation context.
- Use a fresh `request_id` for each query. The envelope still requires `idempotency_key`, but queries do not create idempotency records.
- Treat Evidence content as sensitive. Retrieve it only when metadata and normalized observations cannot answer the question.

## Workflow

1. For a Task, call `GetTask`. If it is running, call `WatchTask` and resume from the last accepted sequence after reconnecting.
2. For a completed Task, prefer `GetTaskReport`; it is the deterministic bundle for its Scope, assets, observations, services, certificates, websites, findings, and Evidence.
3. Use the narrow search query when only one projection is needed: `SearchAssets`, `SearchServices`, `SearchCertificates`, `SearchWebsites`, `SearchFindings`, or `SearchObservations`.
4. Call `GetEvidence` only for referenced evidence IDs that materially affect the conclusion. Verify the returned SHA-256 before interpreting decoded content.
5. For recurring work, call `SearchSchedules`, then inspect `SearchAssetChanges` and `SearchExposureChanges`. Distinguish a first-run baseline from a later run with no change.
6. Call `SearchAudit` only when provenance is requested or a result is disputed. Filter locally to the relevant resource or request ID.
7. Separate confirmed findings, negative evaluations with complete coverage, collection errors, and unknown coverage in the final explanation.

Example query:

```json
{"request_id":"req_...","idempotency_key":"query_...","agent_id":"agent_...","skill_name":"cyberedge-inspect-results","skill_version":"0.1.0","action":"get_task_report","task_id":"task_..."}
```

Set `CYBEREDGE_RPC_SOCKET` only when the service does not use `/tmp/cyberedge.sock`.

## Interpretation rules

- A Finding is supported only when its `observation_id` and `evidence_id` resolve within the same result chain.
- A missing observation is not proof of absence unless the adapter emitted successful coverage for that target.
- `failed` and `canceled` Tasks do not establish a clean baseline.
- Technology hints without an evidence-bound fingerprint are not verified product identities.
- NVD association and public-code references are review candidates, not proof of exploitability or credential exposure.
- Never emit raw binary Evidence or large base64 bodies to the human. Summarize the relevant fact and retain the evidence ID and digest.

## Failure handling

- On `CAPABILITY_DENIED`, report the missing read capability; never switch identity or use a mutation Skill as a workaround.
- On `TASK_NOT_FOUND`, `SCOPE_NOT_FOUND`, or `EVIDENCE_NOT_FOUND`, ask for the correct stable ID.
- Retry only when the typed error sets `retryable=true`.
- If a report is requested before Task completion, watch the Task instead of guessing partial results.

## Output

Return the inspected IDs, Task state, evidence-backed facts, Finding severity/state, monitor changes, explicit coverage gaps, and relevant audit provenance. Cite Evidence by ID and SHA-256. Keep operational errors separate from security conclusions.
