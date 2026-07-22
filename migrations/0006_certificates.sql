CREATE TABLE certificates (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    sha256 TEXT NOT NULL,
    subject TEXT NOT NULL,
    issuer TEXT NOT NULL,
    dns_names JSONB NOT NULL,
    not_before_seconds BIGINT NOT NULL,
    not_before_nanos INTEGER NOT NULL,
    not_after_seconds BIGINT NOT NULL,
    not_after_nanos INTEGER NOT NULL,
    first_seen_at_seconds BIGINT NOT NULL,
    first_seen_at_nanos INTEGER NOT NULL,
    last_seen_at_seconds BIGINT NOT NULL,
    last_seen_at_nanos INTEGER NOT NULL,
    UNIQUE (service_id, sha256)
);

CREATE INDEX certificates_service_id_idx ON certificates (service_id);
