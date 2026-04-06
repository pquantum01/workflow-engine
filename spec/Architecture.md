# Architecture - Workflow Engine

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     React Frontend                       │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────────┐ │
│  │  Admin    │  │  Kanban  │  │  Process Instance      │ │
│  │  Portal   │  │  Board   │  │  Detail View           │ │
│  └──────────┘  └──────────┘  └────────────────────────┘ │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP REST API
┌────────────────────────┴────────────────────────────────┐
│                   Go Backend (API Server)                │
│                                                         │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │  HTTP       │  │  Service      │  │  State Machine │ │
│  │  Handlers   │──│  Layer        │──│  Engine        │ │
│  └─────────────┘  └──────────────┘  └────────────────┘ │
│                          │                              │
│                   ┌──────┴───────┐                      │
│                   │  Repository  │                      │
│                   │  Layer       │                      │
│                   └──────┬───────┘                      │
└──────────────────────────┼──────────────────────────────┘
                           │
                    ┌──────┴───────┐
                    │  PostgreSQL  │
                    └──────────────┘
```

## Monorepo Structure

```
workflow-engine/
├── CLAUDE.md
├── Architecture.md
├── Requirement.md
├── Currentstatus.md
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/             # HTTP handlers (one file per resource)
│   │   │   ├── middleware/           # Auth, logging, CORS, error handling
│   │   │   └── router.go            # Route registration
│   │   ├── model/                    # Domain models (structs)
│   │   ├── service/                  # Business logic layer
│   │   ├── repository/              # Database access layer
│   │   │   └── postgres/            # PostgreSQL implementations
│   │   ├── engine/                   # State machine engine
│   │   └── config/                   # App configuration
│   ├── migrations/                   # SQL migration files
│   ├── tests/
│   │   └── integration/             # Integration tests
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/              # Reusable UI components
│   │   ├── pages/                   # Page-level components
│   │   ├── hooks/                   # Custom React hooks
│   │   ├── api/                     # API client functions
│   │   ├── types/                   # TypeScript type definitions
│   │   └── App.tsx
│   ├── package.json
│   └── tsconfig.json
├── docker-compose.yml                # PostgreSQL + backend for local dev
└── Makefile                          # Build, test, run commands
```

## Backend Architecture

### Layered Architecture

```
HTTP Request
    │
    ▼
┌─────────────────┐
│  Handler Layer   │  Parses HTTP, validates input, returns responses.
│  (api/handler)   │  No business logic here.
└────────┬────────┘
         │ calls
         ▼
┌─────────────────┐
│  Service Layer   │  Business logic, orchestration, state machine
│  (service)       │  evaluation, completion rule checks.
└────────┬────────┘
         │ calls
         ▼
┌─────────────────┐
│  Repository      │  Data access. Interface-based for testability.
│  (repository)    │  PostgreSQL implementation behind interfaces.
└────────┬────────┘
         │
         ▼
     PostgreSQL
```

### Key Packages

| Package | Responsibility |
|---------|---------------|
| `cmd/server` | Application entry point, wiring dependencies |
| `internal/api/handler` | HTTP handlers — parse requests, call services, write responses |
| `internal/api/middleware` | Cross-cutting: logging, CORS, error recovery |
| `internal/model` | Domain structs: ProcessDefinition, WorkflowDefinition, TaskDefinition, StateMachine, and their instance counterparts |
| `internal/service` | Business logic: creation, validation, state transitions, completion rule evaluation |
| `internal/repository` | Interfaces for data access |
| `internal/repository/postgres` | PostgreSQL implementations of repository interfaces |
| `internal/engine` | State machine engine: evaluates transitions, checks guards, applies completion rules |
| `internal/config` | Configuration loading (env vars, config files) |

### State Machine Engine

The engine is the core of the system. It:

1. Loads the state machine definition for a workflow.
2. Evaluates whether a transition is valid from the current state.
3. Checks **completion rules** (are the required tasks done?).
4. Applies the transition atomically.

```go
// Pseudocode
type Engine interface {
    // Evaluate checks if any transitions are now valid and applies them
    Evaluate(ctx context.Context, workflowInstanceID uuid.UUID) (*TransitionResult, error)

    // TriggerTransition manually triggers a named transition
    TriggerTransition(ctx context.Context, workflowInstanceID uuid.UUID, transitionName string) (*TransitionResult, error)
}
```

### Completion Rules

Task-to-state-transition mapping is configured per workflow definition:

```json
{
  "transition": "review_to_approved",
  "rule": {
    "type": "all_of",
    "task_refs": ["code_review", "qa_review"]
  }
}
```

Supported rule types:
- `single`: One specific task must complete.
- `all_of`: All listed tasks must complete.
- `n_of_m`: At least N of the listed tasks must complete.
- `any_of`: At least one of the listed tasks must complete.

## Data Model

### Entity Relationship

```
ProcessDefinition (1) ──── (*) WorkflowDefinition
       │                            │
       │                            │
       │ instantiates               │ instantiates
       ▼                            ▼
ProcessInstance (1) ──────── (*) WorkflowInstance
       │                            │
       │ parent_id (self-ref)       │
       │ (1 level only)             │
       ▼                            ▼
ProcessInstance (children)    TaskInstance (*)
                                    │
                                    │ defined by
                                    ▼
                              TaskDefinition
```

### Core Tables

```sql
-- Process Definitions (templates)
CREATE TABLE process (
    id              UUID PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    version         INT NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, active, archived
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workflow Definitions (belong to a process definition)
CREATE TABLE workflow (
    id                    UUID PRIMARY KEY,
    process_id  UUID NOT NULL REFERENCES process(id),
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    state_machine         JSONB NOT NULL,   -- {states, transitions, initial_state, terminal_states}
    sort_order            INT NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Task Definitions (belong to a workflow definition)
CREATE TABLE task (
    id                      UUID PRIMARY KEY,
    workflow_id             UUID NOT NULL REFERENCES workflow(id),
    name                    VARCHAR(255) NOT NULL,
    description             TEXT,
    task_type               VARCHAR(50) NOT NULL DEFAULT 'manual', -- manual, automated, approval
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Process Instances (running processes)
CREATE TABLE process_instances (
    id                    UUID PRIMARY KEY,
    process_id            UUID NOT NULL REFERENCES process(id),
    parent_id             UUID REFERENCES process_instances(id),  -- 1-level hierarchy
    status                VARCHAR(20) NOT NULL DEFAULT 'running', -- running, completed, cancelled, failed
    started_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_no_grandchildren CHECK (parent_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM process_instances p WHERE p.id = parent_id AND p.parent_id IS NOT NULL
    ))
);

-- Workflow Instances
CREATE TABLE workflow_instances (
    id                      UUID PRIMARY KEY,
    workflow_id             UUID NOT NULL REFERENCES workflow(id),
    process_instance_id     UUID NOT NULL REFERENCES process_instances(id),
    current_state           VARCHAR(100) NOT NULL,
    status                  VARCHAR(20) NOT NULL DEFAULT 'active', -- active, completed, cancelled
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Task Instances
CREATE TABLE task_instances (
    id                     UUID PRIMARY KEY,
    task_id                UUID NOT NULL REFERENCES task(id),
    workflow_instance_id   UUID NOT NULL REFERENCES workflow_instances(id),
    status                 VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, in_progress, completed, skipped, failed
    assigned_to            VARCHAR(255),
    completed_at           TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- State Transition Log (audit trail)
CREATE TABLE state_transitions (
    id                    UUID PRIMARY KEY,
    workflow_instance_id  UUID NOT NULL REFERENCES workflow_instances(id),
    from_state            VARCHAR(100) NOT NULL,
    to_state              VARCHAR(100) NOT NULL,
    transition_name       VARCHAR(255),
    triggered_by          VARCHAR(255),  -- user ID, system, agent, process ID
    trigger_type          VARCHAR(50) NOT NULL, -- human, ai_agent, rule, process
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## API Design

### REST Endpoints

#### Process Definitions
| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/v1/processes` | Create process definition |
| GET    | `/api/v1/processes` | List process definitions |
| GET    | `/api/v1/processes/:id` | Get process definition |
| PUT    | `/api/v1/processes/:id` | Update process definition |
| DELETE | `/api/v1/processes/:id` | Delete process definition |

#### Workflow Definitions
| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/v1/processes/:id/workflows` | Create workflow definition |
| GET    | `/api/v1/processes/:id/workflows` | List workflow definitions |
| GET    | `/api/v1/workflows/:id` | Get workflow definition |
| PUT    | `/api/v1/workflows/:id` | Update workflow definition |
| DELETE | `/api/v1/workflows/:id` | Delete workflow definition |

#### Task Definitions
| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/v1/workflows/:id/tasks` | Create task definition |
| GET    | `/api/v1/workflows/:id/tasks` | List task definitions |
| GET    | `/api/v1/tasks/:id` | Get task definition |
| PUT    | `/api/v1/tasks/:id` | Update task definition |
| DELETE | `/api/v1/tasks/:id` | Delete task definition |

#### Process Instances
| Method | Path | Description |
|--------|------|-------------|
| POST   | `/api/v1/process-instances` | Start a new process instance |
| GET    | `/api/v1/process-instances` | List process instances (filterable) |
| GET    | `/api/v1/process-instances/:id` | Get process instance detail |
| POST   | `/api/v1/process-instances/:id/cancel` | Cancel a process instance |
| GET    | `/api/v1/process-instances/:id/children` | List child process instances |

#### Workflow Instances
| Method | Path | Description |
|--------|------|-------------|
| GET    | `/api/v1/process-instances/:id/workflows` | List workflow instances for a process |
| GET    | `/api/v1/workflow-instances/:id` | Get workflow instance detail |
| POST   | `/api/v1/workflow-instances/:id/transition` | Trigger a manual transition |

#### Task Instances
| Method | Path | Description |
|--------|------|-------------|
| GET    | `/api/v1/workflow-instances/:id/tasks` | List task instances |
| GET    | `/api/v1/task-instances/:id` | Get task instance detail |
| PUT    | `/api/v1/task-instances/:id/status` | Update task status (complete, fail, skip) |
| PUT    | `/api/v1/task-instances/:id/assign` | Assign task to user/agent |

## Technology Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Backend language | Go 1.22+ | Required by spec |
| HTTP router | net/http + gorilla/mux | Standard library server with mux for routing |
| Logging | log/slog | Structured logging, pluggable handlers (JSON, Loggly, etc.) |
| Database | PostgreSQL 16 | JSONB for state machines, strong consistency for state transitions |
| Migrations | golang-migrate | SQL-based, reversible migrations |
| Testing | Go stdlib + testify | Standard assertions + mocking |
| Frontend | React 18 + TypeScript | Required by spec |
| Frontend build | Vite | Fast dev server and build |
| UI library | TBD (likely Ant Design or similar) | Kanban, forms, tables |
| Containerization | Docker Compose | Local dev: PostgreSQL + backend |

## Testing Strategy

### Backend
- **Unit tests**: Every service method, engine rule, and model validation. Repository layer mocked via interfaces.
- **Integration tests**: HTTP handlers tested with a real PostgreSQL database (via Docker). Full request-response cycle.
- **Test database**: Separate PostgreSQL instance (or schema) spun up for integration tests.

### Frontend
- **Unit tests**: Component rendering, hooks, utility functions.
- **Integration tests**: Page-level tests with mocked API responses.

## Development Phases

**Backend-first**: Complete and test all backend functionality before starting frontend.

### Phase 1: Backend Foundation
- Go module initialization, Docker Compose (PostgreSQL), Makefile.
- Structured logging setup (slog with JSON handler).
- Database migrations (all core tables).
- Domain model structs.
- Repository interfaces + PostgreSQL implementations.
- Unit tests for models and repositories.

### Phase 2: Core Engine + Services
- State machine engine (transition evaluation, guard checks).
- Completion rule evaluation (single, all_of, n_of_m, any_of).
- Service layer for definitions (Process, Workflow, Task CRUD).
- Service layer for instances (start process, complete tasks, trigger transitions).
- Parent-child process hierarchy with 1-level enforcement.
- Unit tests for engine and all services.

### Phase 3: API Layer + Integration Tests
- HTTP handlers for all definition CRUD (net/http + gorilla/mux).
- HTTP handlers for instance lifecycle (start, transition, complete task, cancel).
- Middleware (structured logging, error handling, CORS).
- Integration tests for all endpoints (real PostgreSQL via Docker).
- Full API validation and testing — backend is complete and verified.

### Phase 4: Frontend Setup + Admin Portal
- React + TypeScript app scaffolding (Vite).
- API client layer.
- Process/Workflow/Task definition management pages.
- State machine configuration UI.
- Component unit tests.

### Phase 5: Frontend Operations + Polish
- Kanban board for process instances.
- Process instance detail view (workflows, tasks, state transitions).
- Task management UI.
- Webhook notifications.
- Audit log UI.
