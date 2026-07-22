CREATE TABLE findings (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    evidence_id TEXT NOT NULL REFERENCES evidence(id) ON DELETE RESTRICT,
    detector TEXT NOT NULL,
    rule_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    severity INTEGER NOT NULL CHECK (severity BETWEEN 1 AND 5),
    state INTEGER NOT NULL CHECK (state IN (1, 2)),
    fingerprint TEXT NOT NULL,
    first_seen_at_seconds BIGINT NOT NULL,
    first_seen_at_nanos INTEGER NOT NULL,
    last_seen_at_seconds BIGINT NOT NULL,
    last_seen_at_nanos INTEGER NOT NULL,
    UNIQUE (scope_id, detector, rule_id, asset_id, fingerprint)
);

CREATE INDEX findings_scope_severity_idx ON findings (scope_id, severity DESC, last_seen_at_seconds DESC);
