DROP TABLE IS EXISTS task_group;

DROP INDEX IF EXISTS idx_tasks_group_id;

ALTER TABLE tasks DROP COLUMN group_id;