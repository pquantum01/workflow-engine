DROP TABLE IF EXISTS task_attachment;
DROP TABLE IF EXISTS task_comment;
ALTER TABLE task_instance
  DROP COLUMN IF EXISTS is_adhoc,
  DROP COLUMN IF EXISTS expected_completion_time,
  DROP COLUMN IF EXISTS expected_completion_date,
  DROP COLUMN IF EXISTS priority;
