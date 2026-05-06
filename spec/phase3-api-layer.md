# Phase 3: API Layer + Integration Tests

## Overview

Wire up HTTP handlers using `net/http` + `gorilla/mux`, add middleware (logging, CORS, error recovery), and write integration tests that hit real PostgreSQL.

## Dependencies

```
github.com/gorilla/mux  — HTTP router with path params
```

## Package Structure

```
backend/internal/api/
├── router.go              # NewRouter() — registers all routes
├── response.go            # Shared JSON response helpers
├── handler/
│   ├── process.go         # Process definition CRUD
│   ├── workflow.go        # Workflow definition CRUD
│   ├── task.go            # Task definition CRUD
│   ├── process_instance.go # Process instance lifecycle
│   ├── workflow_instance.go # Workflow instance operations
│   └── task_instance.go   # Task instance operations
└── middleware/
    ├── logging.go         # Request/response logging via slog
    ├── recovery.go        # Panic recovery
    └── cors.go            # CORS headers
```

## Handler Pattern

All handlers follow the same pattern:

```go
type ProcessHandler struct {
    service *service.ProcessService
    logger  *slog.Logger
}

func (h *ProcessHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body / path params
    // 2. Call service
    // 3. Write JSON response (201 for create, 200 for get/update, 204 for delete)
    // 4. On error: write error JSON with appropriate status code
}
```

**Thin handlers** — no business logic, just HTTP parsing and response writing.

## Response Format

### Success
```json
{"id": "...", "name": "...", ...}          // single resource
[{"id": "...", "name": "..."}, ...]        // list
```

### Error
```json
{"error": "human-readable error message"}
```

## Shared Response Helpers (response.go)

```go
func writeJSON(w http.ResponseWriter, status int, data any)
func writeError(w http.ResponseWriter, status int, msg string)
```

## Error Mapping

| Service Error | HTTP Status |
|---|---|
| `repository.ErrNotFound` | 404 Not Found |
| Validation error (`"validation:"`) | 400 Bad Request |
| Business rule error (e.g., "cannot update archived") | 409 Conflict |
| JSON decode error | 400 Bad Request |
| Invalid UUID in path | 400 Bad Request |
| Other / unexpected | 500 Internal Server Error |

## Routes (from Architecture.md)

### Process Definitions
| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/processes` | ProcessHandler.Create |
| GET | `/api/v1/processes` | ProcessHandler.List |
| GET | `/api/v1/processes/{id}` | ProcessHandler.GetByID |
| PUT | `/api/v1/processes/{id}` | ProcessHandler.Update |
| DELETE | `/api/v1/processes/{id}` | ProcessHandler.Delete |

### Workflow Definitions
| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/processes/{id}/workflows` | WorkflowHandler.Create |
| GET | `/api/v1/processes/{id}/workflows` | WorkflowHandler.ListByProcessID |
| GET | `/api/v1/workflows/{id}` | WorkflowHandler.GetByID |
| PUT | `/api/v1/workflows/{id}` | WorkflowHandler.Update |
| DELETE | `/api/v1/workflows/{id}` | WorkflowHandler.Delete |

### Task Definitions
| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/workflows/{id}/tasks` | TaskHandler.Create |
| GET | `/api/v1/workflows/{id}/tasks` | TaskHandler.ListByWorkflowID |
| GET | `/api/v1/tasks/{id}` | TaskHandler.GetByID |
| PUT | `/api/v1/tasks/{id}` | TaskHandler.Update |
| DELETE | `/api/v1/tasks/{id}` | TaskHandler.Delete |

### Process Instances
| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/process-instances` | ProcessInstanceHandler.Start |
| GET | `/api/v1/process-instances` | ProcessInstanceHandler.List |
| GET | `/api/v1/process-instances/{id}` | ProcessInstanceHandler.GetByID |
| POST | `/api/v1/process-instances/{id}/cancel` | ProcessInstanceHandler.Cancel |
| GET | `/api/v1/process-instances/{id}/children` | ProcessInstanceHandler.ListChildren |

### Workflow Instances
| Method | Path | Handler |
|---|---|---|
| GET | `/api/v1/process-instances/{id}/workflows` | WorkflowInstanceHandler.ListByProcessInstanceID |
| GET | `/api/v1/workflow-instances/{id}` | WorkflowInstanceHandler.GetByID |
| POST | `/api/v1/workflow-instances/{id}/transition` | WorkflowInstanceHandler.TriggerTransition |

### Task Instances
| Method | Path | Handler |
|---|---|---|
| GET | `/api/v1/workflow-instances/{id}/tasks` | TaskInstanceHandler.ListByWorkflowInstanceID |
| GET | `/api/v1/task-instances/{id}` | TaskInstanceHandler.GetByID |
| PUT | `/api/v1/task-instances/{id}/status` | TaskInstanceHandler.UpdateStatus |
| PUT | `/api/v1/task-instances/{id}/assign` | TaskInstanceHandler.Assign |

## Middleware

### Logging Middleware
- Logs every request: method, path, status code, duration
- Uses `slog` at Info level

### Recovery Middleware
- Catches panics, logs stack trace, returns 500

### CORS Middleware
- Allow all origins in dev (configurable later)
- Allow methods: GET, POST, PUT, DELETE, OPTIONS
- Allow headers: Content-Type, Authorization
- Handle preflight OPTIONS

## Router Setup (router.go)

```go
func NewRouter(
    processService *service.ProcessService,
    workflowService *service.WorkflowService,
    taskService *service.TaskService,
    processInstanceService *service.ProcessInstanceService,
    workflowInstanceService *service.WorkflowInstanceService,
    taskInstanceService *service.TaskInstanceService,
    logger *slog.Logger,
) *mux.Router
```

Registers all routes, wraps with middleware chain: Recovery → Logging → CORS → Handler.

## main.go Updates

Wire up all repos → engine → services → router. Replace the existing `http.NewServeMux()` with the gorilla/mux router.

## Test Plan

### Unit Tests (handler tests with mock services)
Not planned — handlers are thin. Service layer is already well-tested.

### Integration Tests (backend/tests/integration/)
Full HTTP request → response cycle against real PostgreSQL.

| Test | What it verifies |
|---|---|
| Process CRUD | Create, get, list, update, delete; archived blocks update; only draft deletable |
| Workflow CRUD | Create under process, list by process, update, delete; archived process blocks create |
| Task CRUD | Create under workflow, list by workflow, update, delete |
| Process Instance lifecycle | Start, get, list, cancel cascade, parent-child, grandchild rejection |
| Workflow Instance | List by process instance, get, trigger transition |
| Task Instance | List by workflow instance, get, update status, assign, auto-evaluate on completion |
| Error responses | 404 for missing resources, 400 for bad input, 409 for business rule violations |
| Middleware | Logging output, CORS headers, panic recovery |
