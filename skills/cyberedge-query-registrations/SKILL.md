---
name: cyberedge-query-registrations
description: Query normalized ICP registration and organization relationship intelligence for exact domain targets in an explicitly authorized CyberEdge Scope through a configured licensed provider. Use when an AI operator must inspect registration coverage, correlate a domain with an Organization Asset, or report provider-backed relationships without scraping official query pages or exposing personal data.
---

# CyberEdge Query Registrations

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing Scope with a non-empty human authorization reference and domain targets.
- Require a contract-compatible licensed provider and this exact Skill/version grant with `scan.intelligence`, `task.read`, `asset.read`, `evidence.read`, and `report.read`.
- Do not use this Skill to bypass CAPTCHA, access controls, provider licensing, or government sharing approvals.

## Workflow

1. Start `policy_registration_intelligence` for the authorized Scope.
2. Watch the Task until it becomes terminal.
3. Read `registration.icp` Observations for domain-linked registration records.
4. Read `organization.registration` Observations and Organization Assets for normalized enterprise, government, or institution identities.
5. Report ICP number, site name, entity type/name, unified social credit code when present, status, approval date, provider source, Evidence ID, and bounded relationships.
6. Treat `registration.coverage` as completed provider coverage for that exact domain.
7. Treat `registration.error` or `registration.no_targets` as unknown coverage. Do not claim that no registration exists.

Individual registrations are intentionally redacted before Evidence storage. Never request or infer names, phone numbers, addresses, identity numbers, legal representatives, or other personal data. A provider-reported relationship is intelligence provenance, not proof of ownership beyond the stated confidence.

Never invoke the provider directly, widen the query beyond Scope domains, scrape MIIT or public credit-system pages, or submit organization names as caller-controlled lookup terms.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-query-registrations","skill_version":"0.1.0","action":"start_scan","scope_id":"scope-id","policy_id":"policy_registration_intelligence"}
```
