---
name: cyberedge-observe-websites
description: Collect bounded HTTP metadata, response evidence, and verified technology fingerprints from Web services in an explicitly authorized CyberEdge scope.
---

# CyberEdge Observe Websites

Use `cyberedge-agent`; this machine bridge is not a human CLI.

## Preconditions

- Require an existing Scope with a non-empty authorization reference.
- Require this exact Skill/version grant with `scan.active` and `website.read`.
- Do not supply ports, arbitrary URLs, headers or redirect destinations. The server derives endpoints from authorized Scope targets and its fixed baseline.

## Workflow

1. Start `policy_service_baseline` for the authorized Scope.
2. Wait for the Task to become terminal.
3. Call `search_websites` for the same Scope.
4. Correlate Website records and every `fingerprints[].evidence_id` with `http.response` Observations and TaskReport Evidence.
5. Review `discovered_paths` and correlate `http.crawl` Observations with their retained Evidence.
6. If `screenshot_evidence_id` is present, treat it as an offline rendering of retained HTML, not proof of live browser behavior.
7. Correlate `http-host-collision-v1` Findings with `http.host_collision_check` Observations. Require the comparison Evidence to contain both the direct-address baseline and Host-routed response; do not infer a collision from a different body hash alone.
8. Treat `server` only as a header-derived hint. Treat a structured fingerprint as verified only for its versioned detector rule; do not infer additional products or versions.

For recurring monitoring, create a Schedule with the same active policy and call `search_exposure_changes` using its `schedule_id`. Treat `APPEARED`, `DISAPPEARED`, and `MODIFIED` as deterministic diffs between successful Task snapshots. Collection errors suppress disappearance events.

The server probes only open `80`, `443`, `8080`, and `8443`, never follows redirects, enforces a five-second request timeout and rejects response bodies over 1 MiB. `http.error` means collection failed; it does not prove that no website exists.

Current strong-signature rules identify WordPress and Grafana from retained HTML. Generic product-name text is insufficient. WordPress versions are emitted only when an explicit generator value supplies one.

The crawler is depth one and accepts at most 16 same-origin absolute paths found on the root page. It rejects query strings, scheme-relative/cross-origin links, traversal segments, encoded traversal, and caller-supplied paths. Do not describe this as a general-purpose spider.

Host collision checks run only within an active Scope containing domains and alternate IP assets. A domain's current DNS addresses are excluded, candidates are bounded server-side, and probe failures appear as `http.host_collision_error`; treat those failures as unknown coverage, not a negative result.

```json
{"request_id":"req-1","idempotency_key":"idem-1","agent_id":"codex-main","skill_name":"cyberedge-observe-websites","skill_version":"0.1.0","action":"search_websites","scope_id":"scope-id"}
```
