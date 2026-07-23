# ARL capability map

Sources reviewed on 2026-07-22:

- [Aabyss-Team/ARL](https://github.com/Aabyss-Team/ARL)
- [owl234/ARL-Next](https://github.com/owl234/ARL-Next)

CyberEdge copies neither repository's code, database shape, Web workflow, nor operational architecture. This map preserves the useful product capabilities while applying the AI-only control plane and evidence-first domain model.

## Capability mapping

| Upstream capability | CyberEdge domain | Delivery order | Current state |
|---|---|---:|---|
| Domain and IP inventory | Scope, Asset, Observation, Evidence | 1 | DNS vertical slice complete |
| Passive domain data sources | Scanner Adapter, Observation | 2 | DNS and Certificate Transparency implemented |
| Asset grouping and search | Scope and read model | 2 | Scope and bounded search complete |
| Task policy and lifecycle | Policy, Task, Event | 1 | Passive policy and durable lifecycle complete |
| Scheduled and periodic tasks | Schedule producing normal Task | 2 | Implemented baseline |
| Controlled baseline TCP connect scan | Service inventory and Observation | 3 | Implemented |
| Banner and product identification | Website TechnologyFingerprint | 3 | Header hint plus evidence-bound WordPress and Grafana strong-signature identification implemented |
| TLS certificate collection | Certificate Asset, Evidence, Finding | 3 | Implemented with DER evidence and expired/30-day expiry detectors |
| Website and fingerprint discovery | Website Asset, Observation | 3 | Bounded HTTP metadata and body evidence implemented |
| Crawler and screenshots | Website paths, Observation, Evidence | 3 | Bounded crawler, sandbox-capable host adapter, and isolated no-network container renderer implemented |
| File exposure and host collision checks | Finding and Evidence | 4 | Directory listing, exposed Git HEAD, DS_Store, and bounded Host collision detectors implemented |
| Nuclei and custom PoC execution | Finding scanner adapter | 4 | Isolated signed-template Nuclei adapter implemented; no online PoC editor or arbitrary execution will be added |
| Domain and IP asset change monitoring | Monitor asset baseline and diff | 2 | Implemented |
| Website and service change monitoring | Observation diff | 3 | Implemented with coverage-aware exposure changes |
| GitHub leak and CVE monitoring | Threat intelligence adapters | 4 | Bounded GitHub public-code reference monitoring and exact evidence-backed CPE-to-NVD CVE correlation implemented |
| ICP and enterprise relationship lookup | Organization Asset adapter | 4 | Licensed provider-neutral HTTPS/UDS adapter implemented with explicit coverage and personal-data redaction |
| Notifications | Event sink adapters | 3 | Reliable webhook outbox delivery implemented |
| Dashboard and drill-down | Read-only Web projection | 1-3 | Overview, inventory, Task, Scope, evidence count, and audit complete |
| MCP integration | Skill and machine RPC bridge | 1 | Native Skill plus gRPC/JSON bridge complete; MCP compatibility is optional |

## Deliberate differences

- PostgreSQL replaces MongoDB, RabbitMQ, and Celery until measured scale requires another component.
- AI Skills replace the mutable human dashboard and manually configured task forms.
- All scanner output becomes immutable Evidence before it can support a Finding.
- Active scanning requires a separate capability and policy. Passive authorization never upgrades into active authorization.
- Custom PoCs may be packaged and reviewed as versioned adapters later. CyberEdge will not provide a Web source-code editor or arbitrary execution endpoint.
- Monitor and Schedule are definitions that create normal Tasks. They do not become a second execution engine.
