CREATE TABLE websites (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL UNIQUE REFERENCES services(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    status_code INTEGER NOT NULL CHECK (status_code BETWEEN 100 AND 599),
    title TEXT NOT NULL,
    server TEXT NOT NULL,
    content_type TEXT NOT NULL,
    content_sha256 TEXT NOT NULL,
    first_seen_at_seconds BIGINT NOT NULL,
    first_seen_at_nanos INTEGER NOT NULL,
    last_seen_at_seconds BIGINT NOT NULL,
    last_seen_at_nanos INTEGER NOT NULL
);

CREATE INDEX websites_last_seen_idx ON websites (last_seen_at_seconds DESC, last_seen_at_nanos DESC);
