# Workflow Engine

A standalone workflow engine with configurable state machines, tasks, and transitions. Define workflows as JSON, start instances via API, complete tasks to drive state transitions. Built for plugging real-life scenarios from different services via webhooks.

## What

- **Workflow Definitions**: JSON files describing states, transitions, and task templates
- **State Machine Engine**: Linear, task-bound, and user-picks transition modes
- **REST API**: Full CRUD for workflow defs, instances, tasks, and history
- **Webhooks**: Durable outbox-pattern delivery to external services
- **Frontend**: React SPA for managing definitions and operating instances

## Quickstart

### Prerequisites
- Go 1.22+
- Node 20+
- PostgreSQL 16 (local) or Docker

### Option A: Local dev

```bash
# 1. Start Postgres
make docker-up
make migrate-up
make migrate-up-test

# 2. Run backend
make run

# 3. In another terminal, run frontend dev server
make frontend-dev

# 4. Open http://localhost:5173
```

### Option B: Full stack (Docker — requires docker compose v2)

```bash
# start postgres + app
docker compose -f docker-compose.full.yml up --build

# open http://localhost:8080
```

## Define a Workflow

Add JSON files to `./workflows/`. Example:

```json
{
  "name": "leave-approval",
  "version": 1,
  "description": "Employee leave request approval",
  "start_state": "submitted",
  "end_state": "approved",
  "states": [
    {
      "id": "submitted",
      "label": "Submitted",
      "tasks": [
        { "key": "manager_review", "label": "Manager Review", "type": "approval" }
      ]
    },
    { "id": "approved", "label": "Approved", "tasks": [] }
  ],
  "transitions": [
    { "from_state": "submitted", "to_state": "approved" }
  ]
}
```

Reload after adding files:
```bash
curl -X POST http://localhost:8080/api/v1/admin/reload-defs
```

### Transition modes

| Mode | Config | Behavior |
|------|--------|----------|
| **Linear** | Single outgoing edge, no trigger | All tasks done → transition fires |
| **Task-bound** | Multiple edges, all with `trigger.task_key` | Completing bound task fires its edge; siblings cancelled |
| **User-picks** | Multiple edges, no triggers, exactly 1 task | Caller supplies `next_state` on completion |

## API Examples

```bash
BASE=http://localhost:8080/api/v1

# List workflow definitions
curl $BASE/workflow-defs

# Start an instance
curl -X POST $BASE/workflow-instances \
  -H "Content-Type: application/json" \
  -d '{"workflow_def_id":"<def-id>","created_by":"alice@example.com"}'

# List tasks for an instance
curl $BASE/workflow-instances/<inst-id>/tasks

# Complete a task (linear)
curl -X POST $BASE/task-instances/<task-id>/complete \
  -H "Content-Type: application/json" \
  -d '{"status":"done","completed_by":"alice@example.com"}'

# Complete a task (user-picks — must supply next_state)
curl -X POST $BASE/task-instances/<task-id>/complete \
  -H "Content-Type: application/json" \
  -d '{"status":"done","next_state":"hr_review","completed_by":"alice@example.com"}'

# Get instance history
curl $BASE/workflow-instances/<inst-id>/history
```

## Webhooks

Create `webhooks.json` (see `webhooks.json.example`):

```json
[
  {
    "url": "https://your-service.example.com/hooks",
    "secret": "your-shared-secret",
    "event_types": ["task.created", "state.transitioned", "instance.completed"],
    "workflow_def_names": ["leave-approval"]
  }
]
```

Events: `task.created` `task.completed` `task.cancelled` `state.transitioned` `instance.started` `instance.completed` `instance.cancelled`

Signature: `X-Webhook-Signature: sha256=<hmac-hex>`  
Idempotency: `X-Webhook-Delivery: <outbox-row-id>`  
Retry schedule: immediate, +1m, +5m, +30m, +2h, +12h, dead after 6 attempts.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://workflow:workflow@localhost:5432/workflow?sslmode=disable` | Postgres connection string |
| `LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` |
| `LOG_FORMAT` | `json` | `json` \| `text` |
| `HTTP_ADDR` | `:8080` | HTTP listen address |
| `WORKFLOWS_DIR` | `./workflows` | Directory of workflow JSON files |
| `WEBHOOK_CONFIG` | `./webhooks.json` | Webhook subscriber config file |
| `WEBHOOK_WORKER_INTERVAL` | `5s` | How often the outbox worker polls |
| `WEBHOOK_MAX_ATTEMPTS` | `6` | Max delivery attempts before marking failed |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins |

## Testing

```bash
# Backend unit tests
make test

# Backend integration tests (requires postgres)
make test-integration

# Frontend tests
make frontend-test

# Smoke test (requires running server)
./tests/smoke/smoke.sh http://localhost:8080

# Playwright e2e (requires running backend + vite dev server)
cd tests/e2e && BASE_URL=http://localhost:5173 npx playwright test
```

## Architecture

See [spec/Architecture.md](spec/Architecture.md) for full data model, API spec, engine semantics, and tech stack.

## Status

See [spec/Currentstatus.md](spec/Currentstatus.md) for phase progress and decisions log.
