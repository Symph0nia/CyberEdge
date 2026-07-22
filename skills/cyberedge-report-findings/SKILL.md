---
name: cyberedge-report-findings
description: Report and query evidence-backed security findings from authorized CyberEdge Tasks. Use for versioned scanner adapters that classify existing Observations, never for arbitrary command or PoC execution.
---

# CyberEdge Report Findings

Use `cyberedge-agent`. This bridge is machine-only.

## Evidence boundary

- Report only against a running or completed Task.
- Reference an Observation belonging to that exact Task. The server derives Asset and Evidence IDs; never invent them.
- Use a stable detector name, versioned rule ID, and deterministic fingerprint.
- Severity is `1=INFO`, `2=LOW`, `3=MEDIUM`, `4=HIGH`, or `5=CRITICAL`.
- Do not execute commands, upload templates, edit PoCs, or widen Scope through this Skill.

The server deduplicates by Scope, detector, rule, Asset, and fingerprint. A repeated observation refreshes the existing Finding instead of creating noise.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-report-findings","skill_version":"0.1.0","action":"report_finding","task_id":"task-id","observation_id":"observation-id","detector":"adapter-name","rule_id":"rule-id","title":"Finding title","description":"Evidence-backed description","severity":3,"fingerprint":"stable-fingerprint"}
```

Use `search_findings` with `scope_id` for the resulting read model and `get_task_report` for the complete Evidence chain.
