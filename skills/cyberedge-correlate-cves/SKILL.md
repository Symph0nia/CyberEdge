---
name: cyberedge-correlate-cves
description: Correlate exact, evidence-backed CPE identities already observed by CyberEdge with normalized NVD CVE records. Use when an AI operator must run CVE intelligence for an authorized Scope, interpret CVSS metadata and lifecycle state, or distinguish completed CPE coverage from missing identity or provider failure without guessing products or versions.
---

# CyberEdge Correlate CVEs

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing authorized Scope with Website inventory produced by `policy_service_baseline`.
- Require at least one TechnologyFingerprint containing both an exact `cpe_name` and its `cpe_source` provenance.
- Require this exact Skill/version grant with `scan.intelligence`, `task.read`, `finding.read`, `evidence.read`, and `report.read`.

## Workflow

1. Start `policy_cve_intelligence` for the authorized Scope. Never submit a CPE or CVE query directly.
2. Watch the Task until it becomes terminal.
3. Correlate each `detector=nvd-cve` Finding with its `nvd.cve.result` Observation and normalized JSON Evidence.
4. Report the exact CPE, CPE source, CVE ID, NVD status, CVSS score/vector/severity, Evidence ID, references, and Finding lifecycle state.
5. Treat `nvd.cve.coverage` as completed negative coverage only for that exact CPE and Asset.
6. Treat `nvd.cve.error` or `nvd.cve.no_targets` as unknown coverage. Do not claim the Scope is free of known vulnerabilities.

An NVD association means the observed exact product identity falls within NVD applicability data. It does not prove exploitability, reachability, or successful exploitation. State this distinction in reports.

Never infer or edit CPEs, query broad vendor/product wildcards, invoke NVD directly, fetch proof-of-concept code, or upgrade confidence beyond the retained Evidence.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-correlate-cves","skill_version":"0.1.0","action":"start_scan","scope_id":"scope-id","policy_id":"policy_cve_intelligence"}
```
