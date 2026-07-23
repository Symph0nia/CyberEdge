---
name: cyberedge-monitor-public-code
description: Search bounded GitHub public-code metadata for exact domain names in an explicitly authorized CyberEdge Scope. Use when an AI operator must run public-code intelligence, inspect evidence-backed reference candidates, or distinguish completed negative coverage from provider failure without retrieving source content or claiming confirmed credential leakage.
---

# CyberEdge Monitor Public Code

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing Scope with a non-empty human authorization reference and at least one domain target.
- Require this exact Skill/version grant with `scan.intelligence`, `task.read`, `finding.read`, `evidence.read`, and `report.read`.
- Require the isolated provider adapter. The GitHub token belongs only to that adapter.

## Workflow

1. Start `policy_public_code_intelligence` for the authorized Scope.
2. Watch the Task until it becomes terminal.
3. Correlate each `detector=github-public-code` Finding with its `github.code.reference` Observation and JSON Evidence.
4. Report the queried domain, repository, path, blob SHA, GitHub URL, Evidence ID, and lifecycle state.
5. Treat `github.code.coverage` as a completed negative evaluation for that exact domain.
6. Treat `github.code.error` or `github.code.no_targets` as unknown coverage. Do not resolve or suppress prior candidates from an error.

Every result is an `Info` review candidate. It proves only that GitHub code search returned public repository metadata referencing the domain. It does not prove that a secret, credential, exploit, or sensitive source is present.

Never fetch repository content, clone a returned repository, print or request the provider token, widen the query, search domains outside the Scope, or describe a candidate as a confirmed leak.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-monitor-public-code","skill_version":"0.1.0","action":"start_scan","scope_id":"scope-id","policy_id":"policy_public_code_intelligence"}
```
