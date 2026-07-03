# Task Extensions Design

**Date:** 2026-07-03

## Overview

Extends `task_instance` with priority, expected completion date/time, comments, file attachments, and ad-hoc task creation. Mutability rules enforce field lock-down post-completion.

---

## 1. Data Model

### Migration 004 — task_instance new columns

```sql
ALTER TABLE task_instance
  ADD COLUMN priority                TEXT NOT NULL DEFAULT 'medium'
                                     CHECK (priority IN ('low','medium','high')),
  ADD COLUMN expected_completion_date DATE,
  ADD COLUMN expected_completion_time TIME,
  ADD COLUMN is_adhoc                BOOLEAN NOT NULL DEFAULT FALSE;
```

All new columns are nullable (except `priority` which has a default). Existing rows unaffected.

### Migration 004 — new tables

```sql
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
```

Key constraints:
- `task_comment` has no `deleted_at` — immutable by design, no edit/delete allowed
- `UNIQUE (task_instance_id)` on `task_attachment` — one file per task; upload replaces existing
- `s3_url` stores full internal S3 URL (`https://<bucket>.s3.<region>.amazonaws.com/attachments/<task_id>/<filename>`)

---

## 2. Ad-hoc Tasks

### Creation endpoint

`POST /api/v1/workflow-instances/:id/tasks`

```json
{
  "state_id": "review",
  "label": "Review Legal Docs",
  "task_type": "manual",
  "priority": "high",
  "expected_completion_date": "2026-07-10",
  "expected_completion_time": "14:00",
  "created_by": "alice"
}
```

Server-side logic:
1. Load workflow instance → reject `409 INSTANCE_TERMINAL` if instance status is not `active`
2. Load workflow def
3. Validate `state_id` exists in def states — reject `422 INVALID_STATE` if not
4. Slugify `label` + append 4-char random suffix → `task_key` (e.g., `review_legal_docs_a3f9`)
5. Insert `task_instance` with `is_adhoc = TRUE`
6. `priority` defaults to `medium` if omitted

### Key generation

```
label "Review Legal Docs" → slug "review_legal_docs" → key "review_legal_docs_a3f9"
```

Slugify: lowercase, spaces/special chars → `_`, truncate slug to 40 chars, append `_<4-char-hex-random>`. Def task keys use short slugs (`approve`, `legal_review`) — pattern difference makes collision negligible. UNIQUE constraint `(workflow_instance_id, state_id, task_key)` remains as-is.

### Engine impact

No engine changes required. Transition readiness check queries:

```sql
SELECT COUNT(*) FROM task_instance
WHERE workflow_instance_id = $1
  AND state_id = $2
  AND status NOT IN ('done', 'rejected', 'cancelled')
  AND deleted_at IS NULL
```

Ad-hoc tasks with matching `state_id` are automatically counted. Behavior by state:

| State | Effect |
|-------|--------|
| Current state | Blocks transition immediately |
| Future state | Blocks transition when engine reaches that state |
| Past state | No effect — state already transitioned |

---

## 3. Mutability Rules

### Pre-completion (`status: new` or `in_progress`)

| Field | Editable |
|-------|----------|
| `priority` | ✅ |
| `expected_completion_date` | ✅ |
| `expected_completion_time` | ✅ |
| Add comment | ✅ |
| Upload/replace attachment | ✅ (if S3 flag on) |
| `task_label`, `task_type`, `task_key` | ❌ locked always |

### Post-completion (`status: done`, `rejected`, or `cancelled`)

| Field | Editable |
|-------|----------|
| Add comment | ✅ |
| Upload/replace attachment | ✅ (if S3 flag on) |
| `priority`, `expected_completion_date`, `expected_completion_time` | ❌ locked |
| Everything else | ❌ locked |

### Enforcement

- `PATCH /api/v1/task-instances/:id` — **partial update**: only fields present in the request body are applied; omitted fields unchanged. Service checks `task.IsTerminal()` before applying field updates. Returns `409 TASK_COMPLETED` if terminal and any locked field is in the request body.
- `POST /api/v1/task-instances/:id/comments` — no terminal check. Always allowed.
- `POST /api/v1/task-instances/:id/attachment` — no terminal check. Always allowed (subject to S3 flag).

---

## 4. File Upload (S3 + Feature Flag)

### Configuration

```
S3_ENABLED=false          # feature flag — off by default
S3_BUCKET=                # S3 bucket name
S3_REGION=us-east-1       # AWS region
S3_ACCESS_KEY=            # AWS access key
S3_SECRET_KEY=            # AWS secret key
S3_ENDPOINT=              # optional — MinIO/LocalStack endpoint
```

When `S3_ENABLED=false`, server boots normally with empty/dummy S3 config. All attachment endpoints return:

```json
{ "error": { "code": "FEATURE_DISABLED", "message": "file upload is not enabled" } }
```

### Go storage interface

Defined in `internal/storage/`:

```go
type FileStore interface {
    Upload(ctx context.Context, key, contentType string, r io.Reader) (url string, err error)
    Download(ctx context.Context, url string) (io.ReadCloser, error)
    Delete(ctx context.Context, url string) error
}
```

S3 implementation wired at startup only when `S3_ENABLED=true`. Service layer depends on interface — not on S3 SDK directly.

### Upload flow

`POST /api/v1/task-instances/:id/attachment` — JSON body:

```json
{
  "filename": "contract.pdf",
  "content_type": "application/pdf",
  "data": "<base64-encoded-bytes>"
}
```

Steps:
1. Check S3 flag — return `503 FEATURE_DISABLED` if off
2. Decode base64 `data`
3. Validate decoded size ≤ 2MB (accept base64 body up to ~2.73MB; enforce 2MB on decoded)
4. Upload to S3 at key `attachments/{task_instance_id}/{filename}`
5. Upsert `task_attachment` row (replace if exists)
6. Return `201` with attachment metadata (no S3 URL exposed)

### Download flow

`GET /api/v1/task-instances/:id/attachment`

1. Check S3 flag — `503` if off
2. Load `task_attachment` row — `404` if none
3. Fetch object from S3 using stored `s3_url`
4. Stream bytes to client with headers:
   - `Content-Type: <stored content_type>`
   - `Content-Disposition: attachment; filename="<filename>"`

### Delete flow

`DELETE /api/v1/task-instances/:id/attachment`

1. Check S3 flag — `503` if off
2. Load row — `404` if none
3. Delete from S3
4. Delete `task_attachment` row
5. Return `204`

---

## 5. API Surface

### New endpoints

| Method | Path | Description |
|--------|------|-------------|
| `PATCH` | `/api/v1/task-instances/:id` | Update priority / dates |
| `POST` | `/api/v1/workflow-instances/:id/tasks` | Create ad-hoc task |
| `POST` | `/api/v1/task-instances/:id/comments` | Add comment |
| `GET` | `/api/v1/task-instances/:id/comments` | List comments |
| `POST` | `/api/v1/task-instances/:id/attachment` | Upload/replace file |
| `GET` | `/api/v1/task-instances/:id/attachment` | Download file |
| `DELETE` | `/api/v1/task-instances/:id/attachment` | Remove file |

### Error codes

| Code | HTTP | Meaning |
|------|------|---------|
| `TASK_COMPLETED` | 409 | Task terminal, field update not allowed |
| `INSTANCE_TERMINAL` | 409 | Instance not active, ad-hoc task creation rejected |
| `FEATURE_DISABLED` | 503 | S3 flag off |
| `FILE_TOO_LARGE` | 413 | Decoded size > 2MB |
| `INVALID_STATE` | 422 | state_id not in workflow def |

### Existing task response — extended fields

`GET /api/v1/task-instances/:id` gains:

```json
{
  "priority": "medium",
  "expected_completion_date": "2026-07-10",
  "expected_completion_time": "14:00",
  "is_adhoc": false,
  "comment_count": 3,
  "has_attachment": true
}
```

Comments and attachment content fetched via their own endpoints.

---

## 6. Out of Scope

- Multiple attachments per task (one file per task, replace semantics)
- Comment edit or delete
- Auth/identity system (author = free-text string, same pattern as `created_by`)
- CI/CD pipeline changes
