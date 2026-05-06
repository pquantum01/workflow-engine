CREATE SCHEMA IF NOT EXISTS workflow_engine;
SET search_path TO workflow_engine;

-- Process Definitions (templates)
CREATE TABLE process (
    id              UUID PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    version         INT NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workflow Definitions (belong to a process definition)
CREATE TABLE workflow (
    id              UUID PRIMARY KEY,
    process_id      UUID NOT NULL REFERENCES process(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    state_machine   JSONB NOT NULL,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Task Definitions (belong to a workflow definition)
CREATE TABLE task (
    id              UUID PRIMARY KEY,
    workflow_id     UUID NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    task_type       VARCHAR(50) NOT NULL DEFAULT 'manual',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Process Instances (running processes)
CREATE TABLE process_instances (
    id              UUID PRIMARY KEY,
    process_id      UUID NOT NULL REFERENCES process(id),
    parent_id       UUID REFERENCES process_instances(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'running',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workflow Instances
CREATE TABLE workflow_instances (
    id                  UUID PRIMARY KEY,
    workflow_id         UUID NOT NULL REFERENCES workflow(id),
    process_instance_id UUID NOT NULL REFERENCES process_instances(id) ON DELETE CASCADE,
    current_state       VARCHAR(100) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Task Instances
CREATE TABLE task_instances (
    id                   UUID PRIMARY KEY,
    task_id              UUID NOT NULL REFERENCES task(id),
    workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    status               VARCHAR(20) NOT NULL DEFAULT 'pending',
    assigned_to          VARCHAR(255),
    completed_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- State Transition Log (audit trail)
CREATE TABLE state_transitions (
    id                   UUID PRIMARY KEY,
    workflow_instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    from_state           VARCHAR(100) NOT NULL,
    to_state             VARCHAR(100) NOT NULL,
    transition_name      VARCHAR(255),
    triggered_by         VARCHAR(255),
    trigger_type         VARCHAR(50) NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_workflow_process_id ON workflow(process_id);
CREATE INDEX idx_task_workflow_id ON task(workflow_id);
CREATE INDEX idx_process_instances_process_id ON process_instances(process_id);
CREATE INDEX idx_process_instances_parent_id ON process_instances(parent_id);
CREATE INDEX idx_process_instances_status ON process_instances(status);
CREATE INDEX idx_workflow_instances_process_instance_id ON workflow_instances(process_instance_id);
CREATE INDEX idx_workflow_instances_workflow_id ON workflow_instances(workflow_id);
CREATE INDEX idx_task_instances_workflow_instance_id ON task_instances(workflow_instance_id);
CREATE INDEX idx_task_instances_task_id ON task_instances(task_id);
CREATE INDEX idx_task_instances_status ON task_instances(status);
CREATE INDEX idx_state_transitions_workflow_instance_id ON state_transitions(workflow_instance_id);
