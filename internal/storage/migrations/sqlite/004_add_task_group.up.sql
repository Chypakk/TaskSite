CREATE TABLE IF NOT EXISTS task_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tasks ADD COLUMN group_id INTEGER REFERENCES task_groups(id);

CREATE INDEX IF NOT EXISTS idx_tasks_group_id ON tasks(group_id);