# Workflow Engine — Requirements

## Functional Requirements

### Workflow Definitions
- FR-1: System loads workflow definitions from JSON files in a configurable directory at startup.
- FR-2: System validates each definition on load; invalid files log an error and do not block server boot.
- FR-3: Definitions are versioned by `(name, version)`. Same name + version with identical content = no-op. Different content = reject (bump version).
- FR-4: Definitions are immutable once loaded. Editing = new version.
- FR-5: Operator can trigger reload via `POST /api/v1/admin/reload-defs` without restarting server.
- FR-6: Reload response returns `{ loaded, warnings, errors }`.
- FR-7: Soft-deleted definitions are hidden from default list but preserved in DB.
- FR-8: Deleting a definition is blocked if active (non-terminal) instances reference it.

### Workflow Instances
- FR-9: Client starts an instance by supplying `workflow_def_id` and `created_by`.
- FR-10: Starting an instance creates all task instances for the start state.
- FR-11: Instance has status `active | completed | cancelled`.
- FR-12: Client can cancel an active instance; all open tasks become `cancelled`.
- FR-13: Instance stores a `context` JSONB map for accumulated task outputs (v2 use; reserved in v1).
- FR-14: Instance tracks `current_state_id` (string key into workflow def).

### State Machine & Transitions
- FR-15: Engine derives transition mode per state from config shape (linear / task-bound / user-picks). See Architecture.md for semantics.
- FR-16: Linear state: all tasks done → fire single outgoing transition.
- FR-17: Task-bound state: completing the bound task fires its edge; sibling tasks cancelled.
- FR-18: User-picks state (single task, multiple edges, no triggers): caller supplies `next_state` on task completion; engine validates it ∈ outgoing edges.
- FR-19: Reaching `end_state` sets instance status = `completed`.
- FR-20: Any task marked `rejected` → instance status = `cancelled`; all open tasks cancelled.
- FR-21: Every state transition is recorded in `state_transition` (append-only audit log).

### Tasks
- FR-22: Task instances are spawned when their parent state is entered.
- FR-23: Task status lifecycle: `new → in_progress → done | rejected | cancelled`.
- FR-24: `cancelled` is set by engine only (sibling cancel or instance cancel). Not directly settable by client.
- FR-25: Completing a task on a terminal instance returns `409 INSTANCE_TERMINAL`.
- FR-26: Completing a task with status `done | rejected` on an already-terminal task returns `409 CONFLICT`.
- FR-27: Client can mark a task `in_progress` via `POST /task-instances/{id}/start` (optional UI affordance).
- FR-28: Task instances snapshot `task_label` and `task_type` at spawn (immutable; def version changes don't affect in-flight tasks).

### Task Attributes
- FR-29a: Task instances carry `priority` (`low` | `medium` | `high`), default `medium`.
- FR-29b: Task instances carry optional `expected_completion_date` (DATE) and `expected_completion_time` (TIME, time-of-day).
- FR-29c: `priority`, `expected_completion_date`, `expected_completion_time` are editable while task is non-terminal. Locked once terminal.
- FR-29d: `task_label`, `task_type`, `task_key` are immutable after creation.

### Task Comments
- FR-29e: Any caller can add a comment to any task at any time (including after completion) via `POST /api/v1/task-instances/{id}/comments`.
- FR-29f: Comments are immutable — no edit or delete. Author is a free-text string (same pattern as `created_by`).
- FR-29g: Comment `content` is VARCHAR(2048).
- FR-29h: Comments are fetched via `GET /api/v1/task-instances/{id}/comments` (not embedded in task response).

### Task Attachments (File Upload)
- FR-29i: File upload is gated by `S3_ENABLED` env var (default `false`). When disabled, attachment endpoints return `503 FEATURE_DISABLED`.
- FR-29j: One file per task; uploading a new file replaces the existing one.
- FR-29k: Upload accepts base64-encoded file in JSON body (`filename`, `content_type`, `data`). Decoded size must be ≤ 2MB.
- FR-29l: Files stored in S3. Key pattern: `attachments/{task_instance_id}/{filename}`.
- FR-29m: Download proxied through server (`GET /api/v1/task-instances/{id}/attachment`). S3 URL never exposed to client.
- FR-29n: Attachment can be deleted via `DELETE /api/v1/task-instances/{id}/attachment`.
- FR-29o: Upload and delete allowed both pre- and post-completion.

### Ad-hoc Tasks
- FR-29p: Users can add ad-hoc tasks to any workflow instance via `POST /api/v1/workflow-instances/{id}/tasks`.
- FR-29q: Ad-hoc task requires `state_id` (must exist in workflow def), `label`, `task_type`, `created_by`. Optional: `priority`, `expected_completion_date`, `expected_completion_time`.
- FR-29r: Ad-hoc tasks are rejected if the workflow instance is not `active` (returns `409 INSTANCE_TERMINAL`).
- FR-29s: `task_key` for ad-hoc tasks is auto-generated: slugified label + 4-char hex random suffix (e.g., `review_legal_docs_a3f9`).
- FR-29t: Ad-hoc tasks marked `is_adhoc = true`. They participate in transition readiness: all non-terminal tasks for `current_state_id` (including ad-hoc) must be terminal before transition fires.
- FR-29u: Ad-hoc tasks added to past states do not affect transitions (state already moved). Ad-hoc tasks added to future states block transition when the engine reaches that state.

### Task Types (v1)
- FR-30: Four built-in task types: `approval`, `form`, `checklist`, `manual`.
- FR-31: Unknown task types in a def are rejected at load time.
- FR-32a: `GET /api/v1/task-types` returns the canonical list.

### Webhooks
- FR-32: System fires webhooks on events: `task.created`, `task.completed`, `task.cancelled`, `state.transitioned`, `instance.started`, `instance.completed`, `instance.cancelled`.
- FR-33: Webhook subscribers defined in `webhooks.json` (static file, loaded at startup).
- FR-34: Subscribers can filter by `event_types` and optionally by `workflow_def_name`.
- FR-35: Webhook delivery uses outbox pattern: event written to DB in same transaction as state change; worker delivers asynchronously with retry.
- FR-36: Retry schedule: 1m, 5m, 30m, 2h, 12h, dead (6 attempts max, configurable).
- FR-37: Webhook requests signed with HMAC-SHA256 (`X-Webhook-Signature: sha256=<hex>`).
- FR-38: Idempotency key on payload: outbox row ID (`X-Webhook-Delivery` header).

### Observability
- FR-39: All log output via `log/slog` JSON handler. No `fmt.Println`.
- FR-40: Request ID middleware injects `request_id` into every request log entry.
- FR-41: `GET /healthz` returns `200 OK`.
- FR-42: `GET /readyz` returns `200 OK` when DB ping succeeds, `503` otherwise.

---

## Non-Functional Requirements

- NFR-1: **Simplicity** — no unnecessary abstraction. Business logic in services; handlers thin. No global state beyond DI root.
- NFR-2: **Testability** — every package has unit tests. Integration tests hit real Postgres. E2e tests use Playwright + Chrome.
- NFR-3: **Correctness** — concurrent task completions on the same instance serialized via `SELECT FOR UPDATE` on the instance row.
- NFR-4: **Extensibility** — multi-edge conditions (v2), parallel child workflows (v2), user/auth system (v2), Prometheus metrics (v2) are extension points, not retrofits.
- NFR-5: **Durability** — webhook delivery survives server crash (outbox pattern).
- NFR-6: **Single-tenant** — no auth/multi-tenancy in v1. Trusted network assumed.
- NFR-7: **Single binary + single port** — backend serves both API and embedded frontend SPA. One container, port 8080.
- NFR-8: **Postgres 16** only. No other DB target in v1.

---

## Out of Scope (v1)

- Authentication / authorization
- Multi-tenant isolation
- Rule engine / computed conditions on transitions
- Parallel child workflow instances (fork/join)
- Task assignment to specific users
- User/people registry
- Dynamic workflow definition upload via UI
- Prometheus metrics / distributed tracing
- Workflow versioning migration (auto-upgrade running instances to new version)
- Scheduled / timed task auto-completion
- Task type `form` and `checklist` full schema-driven rendering (v1 = stub)
