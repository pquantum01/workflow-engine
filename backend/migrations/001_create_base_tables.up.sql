CREATE TABLE workflow_def (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    version     INT  NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    definition  JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX workflow_def_name_version_unique
    ON workflow_def (name, version)
    WHERE deleted_at IS NULL;

CREATE TABLE workflow_instance (
    id                UUID PRIMARY KEY,
    workflow_def_id   UUID NOT NULL REFERENCES workflow_def(id),
    current_state_id  TEXT NOT NULL,
    status            TEXT NOT NULL CHECK (status IN ('active','completed','cancelled')),
    context           JSONB NOT NULL DEFAULT '{}',
    started_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at      TIMESTAMPTZ,
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX workflow_instance_def_id_idx ON workflow_instance (workflow_def_id) WHERE deleted_at IS NULL;
CREATE INDEX workflow_instance_status_idx ON workflow_instance (status) WHERE deleted_at IS NULL;

CREATE TABLE task_instance (
    id                    UUID PRIMARY KEY,
    workflow_instance_id  UUID NOT NULL REFERENCES workflow_instance(id),
    state_id              TEXT NOT NULL,
    task_key              TEXT NOT NULL,
    task_label            TEXT NOT NULL,
    task_type             TEXT NOT NULL,
    status                TEXT NOT NULL CHECK (status IN ('new','in_progress','done','rejected','cancelled')),
    outcome               TEXT,
    data                  JSONB,
    created_by            TEXT NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at          TIMESTAMPTZ,
    deleted_at            TIMESTAMPTZ,
    UNIQUE (workflow_instance_id, state_id, task_key)
);

CREATE INDEX task_instance_instance_status_idx
    ON task_instance (workflow_instance_id, status)
    WHERE deleted_at IS NULL;

CREATE TABLE state_transition (
    id                    UUID PRIMARY KEY,
    workflow_instance_id  UUID NOT NULL REFERENCES workflow_instance(id),
    from_state            TEXT NOT NULL,
    to_state              TEXT NOT NULL,
    transitioned_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX state_transition_instance_idx ON state_transition (workflow_instance_id, transitioned_at);
