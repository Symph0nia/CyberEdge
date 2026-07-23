---
name: cyberedge-assess-scope
description: Run one complete, evidence-backed CyberEdge assessment workflow for an explicitly authorized Scope. Use when an AI operator must check runtime readiness, expand domain assets, inventory services and websites, run bounded vulnerability and intelligence stages, follow one Task, and report exact stage coverage without claiming unavailable adapters were completed.
---

# CyberEdge Assess Scope

Use `cyberedge-agent` as the JSON machine bridge. This is the umbrella workflow for a full authorized assessment; humans do not operate the CLI directly.

## Workflow

1. Require an existing `scope_id` whose Scope has a non-empty, human-provided `authorization_ref`. If ownership or authorization is ambiguous, stop before active work.
2. Call `GetReadiness`. Report unavailable optional components before starting, but do not split the workflow into unrelated Tasks.
3. Call `GetScope` and verify every requested target is already inside the authorized Scope.
4. Call `StartAssessment` once:
   - use `standard` for routine inventory with the server-owned baseline ports;
   - use `thorough` only when the human explicitly requests comprehensive scanning. It uses the server-owned TCP 1-1024 profile plus selected high-value ports.
5. Follow `WatchTask` until a terminal event. Use `GetTask` if the stream is interrupted. Do not create parallel duplicate assessments.
6. Call `GetTaskReport`. Treat its `coverage` entries as authoritative for workflow completeness.
7. Summarize assets, services, websites, certificates, and findings. For every stage, preserve `complete`, `partial`, `unavailable`, or `blocked`; a completed Task does not mean every stage had coverage.
8. Retrieve Evidence only for material claims that need inspection. Keep hashes and provenance in the report.

## Safety Boundaries

- Never accept an arbitrary port list, URL, header, template, flag, command, CPE, organization search term, or provider query from the caller.
- Never expand outside the authorized root domains or addresses. Discovered names remain bounded to those roots.
- Never evade target throttling, WAF behavior policy, authentication, or provider limits. Mark the affected stage `blocked` or `partial`.
- Never infer a product version or vulnerability without retained evidence.
- On `CAPABILITY_DENIED`, report the missing capability; do not change identity or fall back to direct shell scanners.
- Do not call the low-level policy Skills to make an unavailable assessment stage look complete.

## Output Contract

Return the Task ID, profile, terminal state, readiness snapshot, per-stage coverage, inventory counts, evidence-backed findings, and explicit gaps. Say “full assessment requested” rather than “full assessment completed” unless every required coverage entry is `complete`.
