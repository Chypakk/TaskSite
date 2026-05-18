CREATE TABLE IF NOT EXISTS task_groups (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tasks ADD COLUMN group_id INTEGER;

-- Внешний ключ отдельно
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_group_id 
    FOREIGN KEY (group_id) REFERENCES task_groups(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_group_id ON tasks(group_id);