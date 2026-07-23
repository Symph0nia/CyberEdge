---
name: cyberedge-scan-vulnerabilities
description: Run and interpret CyberEdge's reviewed, bounded Nuclei vulnerability baseline for an explicitly authorized Scope. Use when an AI operator must execute versioned signed templates, wait for the Task, correlate normalized Findings with immutable scanner Evidence, or distinguish negative coverage from adapter failure.
---

# CyberEdge Scan Vulnerabilities

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing Scope with a non-empty human authorization reference.
- Require this exact Skill/version grant with `scan.vulnerability`, `task.read`, `finding.read`, `evidence.read`, and `report.read`.
- Confirm that the deployment has mounted a reviewed, signed template allowlist. Do not upload, generate, select, or edit templates through CyberEdge.

## Workflow

1. Start `policy_vulnerability_baseline` for the authorized Scope.
2. Watch the Task until it becomes terminal.
3. Read the Task report and Scope Findings.
4. Correlate each `detector=nuclei` Finding with its `nuclei.result` Observation and NDJSON Evidence.
5. Treat `nuclei.coverage` as a completed negative evaluation for that derived Website target.
6. Treat `nuclei.error` or `nuclei.no_targets` as unknown coverage. Do not claim the Scope is clean.
7. Report the template ID, matcher, matched location, severity, Evidence ID, and Finding lifecycle state without adding unverified exploitability claims.

The server derives at most 64 credential-free HTTP(S) targets from Website observations. The adapter permits only reviewed signed HTTP/SSL templates, disables redirects, Interactsh, automatic updates, raw request/response output, code, headless, fuzzing, and DoS-tagged templates, and enforces fixed rate/concurrency/time limits.

Never invoke Nuclei directly, pass targets or flags, access the template mount, execute PoCs, or fall back to another scanner.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-scan-vulnerabilities","skill_version":"0.1.0","action":"start_scan","scope_id":"scope-id","policy_id":"policy_vulnerability_baseline"}
```
