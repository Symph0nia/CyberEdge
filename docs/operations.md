# Operations

CyberEdge requires PostgreSQL and an explicit Agent capability policy. The default control transport is a Unix Domain Socket. The optional Web binds only when `CYBEREDGE_WEB_BIND` is set.

## Local container deployment

```bash
docker compose up --build -d
curl --fail http://127.0.0.1:8080/api/v1/overview
```

The Compose port is bound to loopback. Put it behind an authenticated reverse proxy before exposing it to a network. The Web server implements GET and static file routes only; mutation methods return `405`.

Run the AI bridge inside the service container so it can access the local socket:

```bash
printf '%s' '{"request_id":"req_health_scope","idempotency_key":"idem_health_scope","agent_id":"codex-main","skill_name":"cyberedge-discover-assets","skill_version":"0.1.0","action":"create_scope","name":"Example","authorization_ref":"authorization:change-1234","targets":[{"kind":"domain","value":"example.com"}]}' \
  | docker compose exec -T cyberedge cyberedge-agent
```

The bridge accepts one JSON envelope on stdin and emits JSON Lines on stdout. It has no interactive mode.

## Active service baseline

`policy_service_baseline` requires the separate `scan.active` capability and an existing Scope with a non-empty authorization reference. It performs TCP connect checks only against the server-owned baseline set `22,25,53,80,110,143,443,445,3306,5432,6379,8080,8443`; RPC callers cannot supply ports or expand the range. Probes run concurrently with a 750 ms per-port timeout. Service names are port-based hints, not banner-verified product identities.

Open `443` and `8443` services receive a bounded TLS handshake with a two-second connect and three-second handshake timeout. The leaf certificate DER is content-addressed Evidence; Subject, Issuer, DNS SAN, validity and SHA-256 fingerprint form the Certificate read model. Verification is deliberately not a collection gate, so expired and self-signed certificates remain observable; collection never marks a certificate trusted.

Open `80`, `443`, `8080`, and `8443` services receive one `GET /` request. Redirects are never followed, requests time out after five seconds, and response bodies above 1 MiB are rejected. The response body is immutable Evidence; status, title, content type, body hash and the untrusted `Server` header form the Website read model. RPC callers cannot supply a URL, path or headers.

Scheduled active baseline Tasks compare Service and Website observation fingerprints with the preceding successful Task. They emit `APPEARED`, `DISAPPEARED`, or `MODIFIED` ExposureChange records and outbox events. Any collection error suppresses disappearance events for that run, preventing an upstream outage from becoming false exposure drift.

Set `CYBEREDGE_WEBHOOK_URL` to enable notification delivery. Only `http` and `https` endpoints are accepted; redirects are disabled and requests time out after ten seconds. `CYBEREDGE_WEBHOOK_BEARER_TOKEN` optionally adds a bearer credential and is never included in payloads. Outbox events are atomically leased, retried with exponential backoff capped at fifteen minutes, and dead-lettered after eight failed attempts. The read-only Web reports pending, delivered, and dead-letter counts.

Finding adapters use `ReportFinding` and require `finding.report`. A Finding must reference an Observation from the same running or completed Task; Asset and Evidence linkage is derived server-side. Findings deduplicate by Scope, detector, rule, Asset, and stable fingerprint. This contract does not expose arbitrary execution, template upload, or an online PoC editor.

The service baseline also runs the built-in `cyberedge-http/http-directory-listing-v1` detector. It requires a successful HTTP response with both a directory-index title and a listing marker, and stores the raw response body as the referenced Evidence. Built-in findings are committed in the same transaction as their Observation and Evidence. A later successful evaluation resolves a missing condition; recurrence reopens the same deterministic Finding. Probe errors do not resolve findings because coverage is unknown.

Successful TLS probes evaluate `cyberedge-tls/tls-certificate-expired-v1` and `cyberedge-tls/tls-certificate-expiring-v1`. The latter uses a fixed 30-day window. Findings are keyed to the service endpoint rather than a certificate hash, so replacing the certificate resolves the existing endpoint condition instead of leaving a stale finding. The retained DER certificate is the supporting Evidence.

The system HTTP adapter additionally probes only `/.git/HEAD` and `/.DS_Store`. Findings require exact Git HEAD or DS_Store magic signatures; a generic `200` response is insufficient. Redirects remain disabled and the one-megabyte response limit still applies. The RPC contract cannot supply paths, so this does not create an arbitrary URL scanner. Secret-bearing paths such as `/.env` are deliberately excluded from retained Evidence.

Website `fingerprints` are structured, versioned detector projections tied to the root response Evidence. Current rules require an explicit WordPress generator or both WordPress asset families, and Grafana boot data plus a build/title marker. The `server` response header remains an unverified hint and is never promoted into this list by itself.

The system crawler follows at most 16 root-page links at depth one. Only same-origin absolute paths without queries, traversal segments, encoded traversal, or control characters are accepted. Fixed exposure-probe paths are excluded from crawler input. Each fetched page becomes an `http.crawl` Observation with bounded response Evidence; failed pages become `http.crawl_error` and do not widen Scope.

When an active Scope contains domains and alternate IP assets, the baseline performs bounded Host collision checks across at most 16 domains and 64 domain/IP/port candidates with concurrency eight. A domain is never tested against its own current DNS addresses. Detection requires a successful text-like response with a plausible body that differs materially from the direct-address baseline; a changed hash alone is insufficient. The comparison Evidence retains both bounded response bodies and their hashes. Successful negative evaluations resolve an existing `http-host-collision-v1` Finding, while `http.host_collision_error` preserves failed coverage without resolving it.

Set `CYBEREDGE_SCREENSHOTS_ENABLED=true` to explicitly enable the screenshot adapter on a sandbox-capable host and optionally set `CYBEREDGE_CHROMIUM_BIN` (default `/usr/bin/chromium`). It renders only the already-retained root HTML from a temporary local file, with the Chromium sandbox intact, JavaScript disabled, a deny-all network resolver rule, and a deny-all CSP except inline styles/data images. It never navigates Chromium to the target URL. Rendering is limited to 15 seconds and 10 MiB of PNG output; the result becomes `http.screenshot` Evidence and `Website.screenshot_evidence_id`. Temporary HTML, PNG, and browser-profile files are removed after each attempt. Rendering failure records `http.screenshot_error` without failing the website observation.

The core container intentionally contains no Chromium. To enable containerized rendering, add the isolated renderer overlay:

```bash
docker compose -f compose.yaml -f compose.screenshots.yaml up -d --build
```

The renderer receives retained HTML over a shared Unix socket. It has no network namespace, no database credentials, a read-only root filesystem, all Linux capabilities dropped, `no-new-privileges`, bounded memory/PIDs, and only temporary browser storage. Chromium's `--no-sandbox` flag is confined to this sidecar because the container itself is the security boundary. The ordinary `compose.yaml` deployment remains core-only with screenshots disabled.

Do not grant `scan.active` to passive discovery Skills. Keep active grants in a separate Skill binding and verify the Scope before invocation.

## Native runtime

Required:

- `DATABASE_URL`
- `CYBEREDGE_AGENT_POLICY`

Optional local transport:

- `CYBEREDGE_RPC_SOCKET`, default `/tmp/cyberedge.sock`

Optional read-only Web:

- `CYBEREDGE_WEB_BIND`, for example `127.0.0.1:8080`
- `CYBEREDGE_WEB_DIST`, default `web/dist`

## Remote mTLS transport

Set `CYBEREDGE_RPC_ADDR` to enable TCP HTTP/2. The server then requires all three PEM files and does not create the local socket:

- `CYBEREDGE_TLS_CERT`
- `CYBEREDGE_TLS_KEY`
- `CYBEREDGE_TLS_CLIENT_CA`

The AI bridge connects remotely when `CYBEREDGE_RPC_ENDPOINT` is set. It also requires:

- `CYBEREDGE_TLS_DOMAIN`
- `CYBEREDGE_TLS_CA`
- `CYBEREDGE_TLS_CERT`
- `CYBEREDGE_TLS_KEY`

Never reuse the Web reverse-proxy certificate authority as the Agent client CA. Rotate Agent identities independently and bind capabilities to the `agent_id` and Skill metadata in `config/agents.toml`.
