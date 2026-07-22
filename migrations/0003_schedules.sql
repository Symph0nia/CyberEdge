CREATE TABLE schedules (
    id TEXT PRIMARY KEY,
    scope_id TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL,
    interval_seconds BIGINT NOT NULL CHECK (interval_seconds >= 60),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at_seconds BIGINT NOT NULL,
    next_run_at_nanos INTEGER NOT NULL,
    last_task_id TEXT REFERENCES tasks(id),
    created_at_seconds BIGINT NOT NULL,
    created_at_nanos INTEGER NOT NULL
);

CREATE INDEX schedules_due_idx
    ON schedules (next_run_at_seconds, next_run_at_nanos)
    WHERE enabled;

ALTER TABLE tasks ADD COLUMN schedule_id TEXT REFERENCES schedules(id);
