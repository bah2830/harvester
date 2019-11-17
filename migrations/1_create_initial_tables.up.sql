CREATE TABLE IF NOT EXISTS settings (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    settings       TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS jira_time_tracking (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    jira_id             INTEGER NOT NULL,
    jira_description    TEXT NOT NULL,
    started_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    stopped_at          DATETIME
);

CREATE INDEX IF NOT EXISTS jira_time_tracking_jira_id ON jira_time_tracking(jira_id);
CREATE INDEX IF NOT EXISTS jira_time_tracking_started_at ON jira_time_tracking(started_at);