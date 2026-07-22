CREATE TABLE scopes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    authorization_ref TEXT NOT NULL,
    created_at_seconds BIGINT NOT NULL,
    created_at_nanos INTEGER NOT NULL
);

CREATE TABLE scope_targets (
    scope_id TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    kind INTEGER NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (scope_id, position),
    UNIQUE (scope_id, kind, value)
);

CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL REFERENCES scopes(id),
    policy_id TEXT NOT NULL,
    state INTEGER NOT NULL,
    created_at_seconds BIGINT NOT NULL,
    created_at_nanos INTEGER NOT NULL,
    updated_at_seconds BIGINT NOT NULL,
    updated_at_nanos INTEGER NOT NULL
);

CREATE TABLE task_events (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    sequence BIGINT NOT NULL CHECK (sequence > 0),
    event_type TEXT NOT NULL,
    occurred_at_seconds BIGINT NOT NULL,
    occurred_at_nanos INTEGER NOT NULL,
    PRIMARY KEY (task_id, sequence)
);

CREATE TABLE idempotency_keys (
    operation TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    request_fingerprint BYTEA NOT NULL,
    resource_kind TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (operation, agent_id, idempotency_key)
);

CREATE TABLE audit_events (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    skill_name TEXT NOT NULL,
    skill_version TEXT NOT NULL,
    resource_kind TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE outbox_events (
    id TEXT PRIMARY KEY,
    aggregate_kind TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    sequence BIGINT,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE INDEX outbox_events_unpublished_idx
    ON outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE INDEX audit_events_actor_idx
    ON audit_events (agent_id, occurred_at DESC);

