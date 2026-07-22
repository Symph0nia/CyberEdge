CREATE TABLE exposure_changes (
    id TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    resource_kind TEXT NOT NULL CHECK (resource_kind IN ('service', 'website')),
    resource_id TEXT NOT NULL,
    kind INTEGER NOT NULL,
    previous_fingerprint TEXT NOT NULL,
    current_fingerprint TEXT NOT NULL,
    detected_at_seconds BIGINT NOT NULL,
    detected_at_nanos INTEGER NOT NULL
);

CREATE INDEX exposure_changes_schedule_idx
    ON exposure_changes (schedule_id, detected_at_seconds DESC, detected_at_nanos DESC);
