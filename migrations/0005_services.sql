CREATE TABLE services (
    id TEXT PRIMARY KEY,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    transport TEXT NOT NULL,
    port INTEGER NOT NULL CHECK (port > 0 AND port <= 65535),
    service_hint TEXT NOT NULL,
    first_seen_at_seconds BIGINT NOT NULL,
    first_seen_at_nanos INTEGER NOT NULL,
    last_seen_at_seconds BIGINT NOT NULL,
    last_seen_at_nanos INTEGER NOT NULL,
    UNIQUE (asset_id, transport, port)
);

CREATE INDEX services_asset_idx ON services (asset_id, port);
