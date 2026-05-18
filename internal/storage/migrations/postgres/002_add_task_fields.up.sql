ALTER TABLE tasks ADD COLUMN description TEXT DEFAULT '';
ALTER TABLE tasks ADD COLUMN author TEXT NOT NULL DEFAULT 'Unknown';
ALTER TABLE tasks ADD COLUMN status TEXT NOT NULL DEFAULT 'open';
ALTER TABLE tasks ADD COLUMN user_id INTEGER;
ALTER TABLE tasks ADD COLUMN updated_at TIMESTAMPTZ;
ALTER TABLE tasks ADD COLUMN completed_at TIMESTAMPTZ;

-- Внешний ключ добавляем отдельно
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_user_id 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_tasks_status ON tasks(status);