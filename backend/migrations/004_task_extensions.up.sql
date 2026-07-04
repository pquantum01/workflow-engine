ALTER TABLE task_instance
  ADD COLUMN priority                TEXT NOT NULL DEFAULT 'medium'
                                     CHECK (priority IN ('low','medium','high')),
  ADD COLUMN expected_completion_date DATE,
  ADD COLUMN expected_completion_time TIME,
  ADD COLUMN is_adhoc                BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE task_comment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL REFERENCES task_instance(id),
  author           TEXT NOT NULL,
  content          VARCHAR(2048) NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX task_comment_task_idx ON task_comment (task_instance_id, created_at);

CREATE TABLE task_attachment (
  id               UUID PRIMARY KEY,
  task_instance_id UUID NOT NULL UNIQUE,
  filename         TEXT NOT NULL,
  content_type     TEXT NOT NULL,
  size_bytes       INT  NOT NULL,
  s3_url           TEXT NOT NULL,
  uploaded_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
