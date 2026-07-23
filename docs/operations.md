# Operations

CyberEdge requires PostgreSQL and an explicit Agent capability policy. The default control transport is a Unix Domain Socket. The optional Web binds only when `CYBEREDGE_WEB_BIND` is set.

## Local container deployment

```bash
docker compose up --build -d
```

The base deployment does not start or publish the optional Web projection.

Run the AI bridge inside the service container so it can access the local socket:

```bash
printf '%s' '{"request_id":"req_health_scope","idempotency_key":"idem_health_scope","agent_id":"codex-main","skill_name":"cyberedge-discover-assets","skill_version":"0.1.0","action":"create_scope","name":"Example","authorization_ref":"authorization:change-1234","targets":[{"kind":"domain","value":"example.com"}]}' \
  | docker compose exec -T cyberedge cyberedge-agent
```

The bridge accepts one JSON envelope on stdin and emits JSON Lines on stdout. It has no interactive mode.

`GetTaskReport` includes Evidence identifiers, media types, hashes, and timestamps but deliberately omits Evidence bodies so comprehensive reports remain bounded. Use `GetEvidence` for the small number of artifacts needed to support a material conclusion.

## End-to-end assessment

The `cyberedge-assess-scope` Skill first calls `GetReadiness`, then starts one `StartAssessment` Task. `standard` uses the fixed baseline service ports. `thorough` uses the server-owned TCP `1-1024` profile plus selected high-value ports; callers still cannot provide ports, templates, flags, headers, paths, commands, or targets outside the Scope.

The Task chains passive DNS and Certificate Transparency, a bounded in-domain DNS label set, active inventory, Host collision comparison, the reviewed Nuclei baseline, exact-domain public-code metadata, exact observed CPE-to-NVD correlation, and licensed registration intelligence. `GetTaskReport.coverage` reports each stage as `complete`, `partial`, `unavailable`, or `blocked`. Treat Task state `COMPLETED` only as “the workflow stopped normally,” never as proof that optional adapters ran.

Certificate Transparency discovery queries `crt.sh` first and falls back to the Cert Spotter issuance API when the primary request fails. The fallback is bounded to one page of unexpired issuances, requests DNS names only, and uses the same root-domain normalization and wildcard filtering as the primary source.

```bash
printf '%s' '{"request_id":"req_assess_1","idempotency_key":"idem_assess_1","agent_id":"codex-main","skill_name":"cyberedge-assess-scope","skill_version":"0.1.0","action":"start_assessment","scope_id":"scope_...","profile":"thorough"}' \
  | docker compose exec -T cyberedge cyberedge-agent
```

## Active service baseline

`policy_service_baseline` requires the separate `scan.active` capability and an existing Scope with a non-empty authorization reference. It performs TCP connect checks only against the server-owned baseline set `22,25,53,80,110,143,443,445,3306,5432,6379,8080,8443`; RPC callers cannot supply ports or expand the range. Probes run concurrently with a 750 ms per-port timeout. Service names are port-based hints, not banner-verified product identities.

Open `443` and `8443` services receive a bounded TLS handshake with a two-second connect and three-second handshake timeout. The leaf certificate DER is content-addressed Evidence; Subject, Issuer, DNS SAN, validity and SHA-256 fingerprint form the Certificate read model. Verification is deliberately not a collection gate, so expired and self-signed certificates remain observable; collection never marks a certificate trusted.

Open `80`, `443`, `8080`, and `8443` services receive one `GET /` request. Redirects are never followed, requests time out after five seconds, and response bodies above 1 MiB are rejected. The response body is immutable Evidence; status, title, content type, body hash and the untrusted `Server` header form the Website read model. RPC callers cannot supply a URL, path or headers.

Scheduled active baseline Tasks compare Service and Website observation fingerprints with the preceding successful Task. They emit `APPEARED`, `DISAPPEARED`, or `MODIFIED` ExposureChange records and outbox events. Any collection error suppresses disappearance events for that run, preventing an upstream outage from becoming false exposure drift.

Set `CYBEREDGE_WEBHOOK_URL` to enable notification delivery. Only `http` and `https` endpoints are accepted; redirects are disabled and requests time out after ten seconds. `CYBEREDGE_WEBHOOK_BEARER_TOKEN` optionally adds a bearer credential and is never included in payloads. Outbox events are atomically leased, retried with exponential backoff capped at fifteen minutes, and dead-lettered after eight failed attempts. The read-only Web reports pending, delivered, and dead-letter counts.

Finding adapters use `ReportFinding` and require `finding.report`. A Finding must reference an Observation from the same running or completed Task; Asset and Evidence linkage is derived server-side. Findings deduplicate by Scope, detector, rule, Asset, and stable fingerprint. This contract does not expose arbitrary execution, template upload, or an online PoC editor.

## Vulnerability baseline

`policy_vulnerability_baseline` requires the separate `scan.vulnerability` capability. It first executes the fixed active inventory baseline, then sends at most 64 successfully observed, credential-free Website origins to the Nuclei adapter over a Unix socket. AI callers cannot provide targets, templates, tags, severity filters, flags, headers, variables, or commands.

Populate `config/nuclei-templates/` with a reviewed, signed allowlist before enabling the adapter, or set `CYBEREDGE_NUCLEI_TEMPLATES_DIR` to an absolute operator-managed directory. Template lifecycle is an operator-controlled release process; the runtime never downloads or updates templates. Start the deployment with:

```bash
docker compose -f compose.yaml -f compose.nuclei.yaml up -d --build
```

The adapter image pins Nuclei v3.11.0 by digest and has no database credentials. It runs on a separate egress network with a read-only root filesystem, all capabilities dropped, `no-new-privileges`, bounded memory/PIDs, and temporary working storage. Its fixed profile permits signed HTTP/SSL templates only; it disables unsigned templates, automatic updates, redirects, Interactsh/OAST, code, headless, fuzz/DAST, DoS tags, stdin, raw request/response output, and embedded template output. Rate is capped at 20 requests/second with concurrency and bulk size five, a five-second request timeout, one retry, and a nine-minute process deadline.

Each accepted JSONL match is schema-checked and normalized before becoming immutable `application/x-ndjson` Evidence: raw request/response, embedded templates and interaction records are rejected, while Nuclei's generated `curl-command` is removed and replaced by the SHA-256 of the original source line. The result creates a `nuclei.result` Observation and a Finding keyed by template ID, matcher, matched location, and Asset. A successful target evaluation emits `nuclei.coverage`, allowing missing prior matches to resolve and later recurrence to reopen the same Finding. Adapter errors, malformed/out-of-scope output, and missing targets emit `nuclei.error` or `nuclei.no_targets` and never resolve existing Findings.

## Public-code intelligence adapter

Write the GitHub token to a deployment secret file readable only by UID `65532`, set `CYBEREDGE_GITHUB_TOKEN_FILE` to that file's absolute path, and layer `compose.public-code.yaml` over `compose.yaml`. Local Docker Compose bind-mounts secret files without changing ownership, so prepare the file as UID/GID `65532` with mode `0400`; do not make it world-readable. Compose mounts it as `/run/secrets/github_token` only inside the isolated `public-code-adapter`; the token is absent from container environment variables, and the core sees only a Unix socket. The sidecar has a read-only root filesystem, no Linux capabilities, a private egress network, bounded memory/processes, and no database network or credentials.

`policy_public_code_intelligence` requires `scan.intelligence`. It derives at most eight exact domain names from the authorized Scope and performs one quoted GitHub code search per domain with at most 20 results. The adapter retains only query domain, public repository, path, file name, blob SHA, and GitHub URL. It deliberately discards response fragments and never fetches source content. Each accepted item creates `github.code.reference` Evidence and an `Info` Finding explicitly classified as a review candidate, while `github.code.coverage` drives resolution. Provider errors and malformed/out-of-scope output never resolve prior candidates.

This is not GitHub Secret Scanning and does not prove a credential leak.

## Exact CPE and NVD CVE adapter

Layer `compose.cve.yaml` over `compose.yaml`. Anonymous NVD access works with the default empty secret; for production rate limits, write an NVD API key to a deployment secret file and set `CYBEREDGE_NVD_API_KEY_FILE` to its absolute path. Compose mounts it read-only only inside `cve-adapter`. The core receives no NVD credentials and communicates through `/run/cyberedge-cve/adapter.sock`.

`policy_cve_intelligence` requires `scan.intelligence`. The core derives at most 16 CPE 2.3 names from persisted Website TechnologyFingerprint records. A CPE candidate is eligible only when vendor, product, and version are exact and `cpe_source` records the versioned CyberEdge mapping rule. The sidecar first requires an exact match from the official NVD CPE Dictionary API and only then queries CVEs. Current strong WordPress generator detection can produce a candidate when an explicit version is present; unversioned Grafana detection deliberately produces no CPE and therefore no provider query.

The adapter queries NVD CVE API 2.0 by exact `cpeName`, then retains only normalized CVE identity, dates/status, English description, preferred CVSS metadata, and at most ten references. Raw configurations and applicability trees are discarded. Each association produces `nvd.cve.result` Evidence and a lifecycle-managed Finding. `nvd.cve.coverage` resolves a missing prior association for that exact CPE and Asset; provider errors or missing CPE identity never resolve prior Findings. An NVD association is not proof that the deployed path is reachable or exploitable.

## Licensed registration intelligence adapter

CyberEdge does not automate CAPTCHA-protected MIIT or enterprise-credit query pages. Configure a contract-compatible licensed provider with `CYBEREDGE_REGISTRATION_PROVIDER_URL`, write its bearer token to a deployment secret file, set `CYBEREDGE_REGISTRATION_PROVIDER_TOKEN_FILE`, and layer `compose.registration.yaml` over `compose.yaml`. The endpoint must use HTTPS. The token is mounted only inside `registration-adapter`; the core sees only a Unix socket.

`policy_registration_intelligence` requires `scan.intelligence` and sends at most eight exact Scope domains. The provider must explicitly return complete coverage for every requested domain. Results become `registration.icp` domain Observations; non-individual entities additionally become stable Organization Assets with `organization.registration` Evidence. Individual entity names and identifiers are cleared inside the adapter. Unknown fields—including phone, address, identity number, legal representative, email, and raw payload fields—are rejected. See [registration-provider-contract.md](registration-provider-contract.md) for the exact schema and failure semantics.

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

The Web has no mutation routes, but read-only does not mean public. Network-bound deployments require all of:

- `CYBEREDGE_WEB_OIDC_ISSUER`
- `CYBEREDGE_WEB_OIDC_AUDIENCE`
- `CYBEREDGE_WEB_OIDC_JWKS_URL`, HTTPS only

CyberEdge validates the Bearer JWT signature against the configured JWKS and requires `exp`, `iss`, `aud`, and `sub`. Only `RS256` is accepted. Redirects are disabled while fetching JWKS, the request times out after ten seconds, and the document is capped at 1 MiB. Keys are cached for fifteen minutes and refreshed when the cache expires or a valid token presents an unknown `kid`; refreshes are serialized to avoid an IdP request stampede.

The role claim defaults to `roles`. `cyberedge.read` grants the read-only UI and inventory APIs; raw Evidence additionally requires `cyberedge.evidence.read`. Scope authorization references and Audit request, Agent, and Skill identities are omitted unless the token also has `cyberedge.sensitive.read`. Override these names with `CYBEREDGE_WEB_ROLE_CLAIM`, `CYBEREDGE_WEB_READ_ROLE`, `CYBEREDGE_WEB_EVIDENCE_ROLE`, and `CYBEREDGE_WEB_SENSITIVE_ROLE`. A reverse proxy may perform the browser login flow, but it must forward the signed token as `Authorization: Bearer`; unsigned identity headers are never trusted.

Start the authenticated container projection with:

```bash
docker compose -f compose.yaml -f compose.web.yaml up -d --build
```

For a native developer preview only, bind to a loopback address and set `CYBEREDGE_WEB_ALLOW_INSECURE_LOCAL=true`. This escape hatch is rejected for wildcard and non-loopback binds.

The observer uses URL-addressable top-level views for overview, inventory, findings, Tasks, monitoring, and audit. Search spans the loaded read projections; Finding severity and Task state filters can be saved locally in the browser without creating server-side state. Asset and Finding inspectors expose relationships, Task inspectors load their Observation timeline, and Evidence content is fetched only on explicit inspection. JSON export is generated from the current read-only snapshot. None of these interactions call a mutation route.

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
