CREATE TABLE IF NOT EXISTS settings (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    settings       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS time_tracking (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    jira_id         INTEGER NOT NULL,
    track_type      TEXT NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS time_tracking_jira_id ON time_tracking(jira_id);
CREATE INDEX IF NOT EXISTS time_tracking_track_type ON time_tracking(track_type);
CREATE INDEX IF NOT EXISTS time_tracking_created_at ON time_tracking(created_at);
CREATE INDEX IF NOT EXISTS time_tracking_jira_all ON time_tracking(jira_id, track_type);