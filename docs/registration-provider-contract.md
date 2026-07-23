# Registration intelligence provider contract

CyberEdge does not scrape MIIT ICP or National Enterprise Credit Information Publicity System query pages. Deployments integrate a licensed or explicitly authorized data provider through one outbound HTTPS endpoint.

## Request

The adapter sends `POST` with `Authorization: Bearer <token>` and `Content-Type: application/json`:

```json
{"domains":["example.cn"]}
```

Only normalized domain targets derived from the authorized Scope are sent. A request contains at most eight domains. Callers cannot supply organization names, credit codes, provider filters, pagination, or query syntax.

## Response

The provider must return one `coverage` entry with `status=complete` for every requested domain. Omitting coverage fails the whole Task safely; absence of a record is not silently treated as a negative result.

```json
{
  "coverage": [{"domain": "example.cn", "status": "complete"}],
  "records": [{
    "domain": "example.cn",
    "icp_number": "京ICP备00000000号-1",
    "site_name": "Example Site",
    "entity_name": "Example Technology Co., Ltd.",
    "entity_type": "enterprise",
    "unified_social_credit_code": "91110000123456789X",
    "status": "active",
    "approved_at": "2026-01-01",
    "source": "licensed-provider",
    "source_url": "https://provider.example/records/1",
    "relationships": [{
      "relationship_type": "affiliate",
      "related_entity": "Example Cloud Co., Ltd.",
      "related_domain": "cloud.example.cn",
      "confidence": "confirmed",
      "source": "licensed-provider"
    }]
  }]
}
```

Allowed entity types are `enterprise`, `government`, `institution`, `individual`, and `other`. Allowed record statuses are `active`, `cancelled`, and `unknown`. Allowed relationships are `same_entity`, `parent`, `subsidiary`, and `affiliate`; confidence is `confirmed` or `reported`.

The schema rejects unknown fields. Phone numbers, addresses, identity numbers, legal representatives, shareholders, emails, and raw provider payloads cannot enter Evidence. For `individual`, the adapter always clears `entity_name` and `unified_social_credit_code` before returning data to the core.

## Failure semantics

- Non-2xx responses, redirects, incomplete coverage, unknown fields, invalid domains, or oversized data become `registration.error`.
- A successful covered domain with no records emits only `registration.coverage`.
- Enterprise, government, and institution records create a stable Organization Asset. Credit code is used as its identity when present; otherwise the normalized entity name is used.
- Provider credentials are mounted as a Docker secret only into the adapter. The CyberEdge core and read-only Web never receive them.
