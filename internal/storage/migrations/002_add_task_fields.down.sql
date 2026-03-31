DROP INDEX IF EXISTS idx_tasks_status;

ALTER TABLE tasks DROP COLUMN completed_at;
ALTER TABLE tasks DROP COLUMN user_id;
ALTER TABLE tasks DROP COLUMN status;
ALTER TABLE tasks DROP COLUMN author;
ALTER TABLE tasks DROP COLUMN description;