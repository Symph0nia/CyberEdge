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
