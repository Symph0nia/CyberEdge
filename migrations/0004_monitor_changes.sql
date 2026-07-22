CREATE TABLE asset_changes (
    id TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    kind INTEGER NOT NULL CHECK (kind IN (1, 2)),
    detected_at_seconds BIGINT NOT NULL,
    detected_at_nanos INTEGER NOT NULL,
    UNIQUE (task_id, asset_id, kind)
);

CREATE INDEX asset_changes_schedule_idx
    ON asset_changes (schedule_id, detected_at_seconds DESC, detected_at_nanos DESC);
