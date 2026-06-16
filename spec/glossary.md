# Glossary

## Core Terms

**Workflow Definition (`workflow_def`)**
A versioned blueprint describing states, transitions, and task templates. Immutable once loaded. Identified by `(name, version)`. Stored as JSONB in DB.

**State**
A named stage within a workflow definition. Has a string `id` (workflow-local, slug format), a `label`, and zero or more task templates. Every non-end state has at least one task and at least one outgoing transition.

**Start State**
The first state entered when a workflow instance is created. Declared as `start_state` in the definition.

**End State**
The single terminal success state. Declared as `end_state`. Has zero tasks and zero outgoing transitions. Reaching it completes the instance.

**Transition**
A directed edge from one state to another. Has `from_state`, `to_state`, and an optional `trigger`.

**Trigger**
A condition on a transition declaring which task key fires it (v1: `trigger.task_key`). Absent trigger = unconditional (used by linear and user-picks states).

**Linear State**
A state with exactly one outgoing transition and no triggers. All tasks must complete (`done`) before the transition fires.

**Task-Bound State**
A state with multiple outgoing transitions where every transition has a `trigger.task_key`. Completing the bound task fires its edge; sibling tasks are cancelled.

**User-Picks State**
A state with multiple outgoing transitions, no triggers, and exactly one task. The caller supplies `next_state` when completing that task. Engine validates the choice.

**Decision State**
Umbrella term for Task-Bound State or User-Picks State — any state where a single task completion determines which path is taken.

**Task Template**
The definition of a task within a state: `key`, `label`, `type`. Blueprint; not an instance.

**Task Instance (`task_instance`)**
A concrete unit of work spawned when its parent state is entered. Carries a snapshot of the template's `label` and `type`. Status: `new → in_progress → done | rejected | cancelled`.

**Workflow Instance (`workflow_instance`)**
A running execution of a workflow definition. Tracks `current_state_id`, `status`, `context`, and timing. Status: `active | completed | cancelled`.

**State Transition (audit record)**
An append-only record written each time an instance moves from one state to another. Feeds the history timeline.

**Engine**
The Go package (`internal/engine`) that owns state machine logic: spawning tasks, evaluating transition mode, firing transitions, managing terminal states.

**Workflow Loader**
The Go package (`internal/workflowloader`) that scans `./workflows/*.json` at startup, validates each file, and upserts into DB.

**Webhook Outbox**
A DB table written in the same transaction as state changes. A background worker reads it and POSTs to subscribers. Guarantees at-least-once delivery.

**`created_by`**
Free-text caller-supplied identifier for who initiated an instance or completed a task. No user system in v1.

**Context (instance)**
A JSONB map on the workflow instance for accumulated task outputs. Reserved for v2 use; empty in v1.

**Outcome (task)**
Optional string set on task completion indicating which branch was chosen (e.g., `"approved"`, `"rejected"`). Reserved for v2 trigger conditions; persisted but not evaluated in v1.

**Soft Delete**
Marking a row `deleted_at = now()` instead of physical removal. Default queries filter `WHERE deleted_at IS NULL`.

**`(name, version)` uniqueness**
Workflow definitions are unique by `(name, version)` ignoring soft-deleted rows. A deleted def's slot can be reused.

## Slug Format
State IDs, task keys, and workflow def names follow `^[a-z][a-z0-9_-]*$`. Enforced at load time. Canonical form required (no auto-lowercase).

## Task Types (v1)
`approval` · `form` · `checklist` · `manual` — Go const set. Unknown types rejected at load.
