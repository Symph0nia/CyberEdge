CREATE TABLE assets (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    kind INTEGER NOT NULL,
    value TEXT NOT NULL,
    first_seen_at_seconds BIGINT NOT NULL,
    first_seen_at_nanos INTEGER NOT NULL,
    last_seen_at_seconds BIGINT NOT NULL,
    last_seen_at_nanos INTEGER NOT NULL,
    UNIQUE (scope_id, kind, value)
);

CREATE TABLE evidence (
    id TEXT PRIMARY KEY,
    media_type TEXT NOT NULL,
    sha256 TEXT NOT NULL UNIQUE,
    content BYTEA NOT NULL,
    created_at_seconds BIGINT NOT NULL,
    created_at_nanos INTEGER NOT NULL
);

CREATE TABLE observations (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    observation_type TEXT NOT NULL,
    value_json JSONB NOT NULL,
    evidence_id TEXT NOT NULL REFERENCES evidence(id),
    observed_at_seconds BIGINT NOT NULL,
    observed_at_nanos INTEGER NOT NULL
);

CREATE INDEX assets_scope_idx ON assets (scope_id, kind, value);
CREATE INDEX observations_task_idx ON observations (task_id, observed_at_seconds);
