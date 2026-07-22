ALTER TABLE outbox_events
    ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN lease_until TIMESTAMPTZ,
    ADD COLUMN last_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN dead_lettered_at TIMESTAMPTZ;

DROP INDEX outbox_events_unpublished_idx;
CREATE INDEX outbox_events_delivery_idx
    ON outbox_events (next_attempt_at, created_at)
    WHERE published_at IS NULL AND dead_lettered_at IS NULL;
