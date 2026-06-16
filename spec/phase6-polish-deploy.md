# Phase 6 — Polish & Deploy

## Goal
Production-ready packaging: Dockerfile, docker-compose full stack, graceful shutdown, smoke tests, Playwright e2e tests, README.

---

## Dockerfile (`backend/Dockerfile`)

Three-stage build: frontend → backend → runtime.

```dockerfile
# Stage 1: Frontend build
FROM node:20-alpine AS web
WORKDIR /web
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build
# outputs: /web/dist

# Stage 2: Go build (embeds frontend dist)
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.work ./
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ ./backend/
COPY --from=web /web/dist ./backend/webroot/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" \
    -o /out/server ./backend/cmd/server

# Stage 3: Runtime
FROM alpine:3.20
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/server /server
COPY backend/migrations /migrations
COPY workflows /workflows
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1
ENTRYPOINT ["/server"]
```

---

## docker-compose.yml (full stack)

```yaml
version: "3.9"
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: workflow
      POSTGRES_USER: workflow
      POSTGRES_PASSWORD: workflow
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "workflow"]
      interval: 5s
      timeout: 3s
      retries: 10

  app:
    build:
      context: .
      dockerfile: backend/Dockerfile
    environment:
      DATABASE_URL: postgres://workflow:workflow@postgres:5432/workflow?sslmode=disable
      LOG_LEVEL: info
      LOG_FORMAT: json
      HTTP_ADDR: :8080
      WORKFLOWS_DIR: /workflows
      WEBHOOK_CONFIG: /run/config/webhooks.json
      CORS_ALLOWED_ORIGINS: "*"
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./webhooks.json:/run/config/webhooks.json:ro   # optional; empty file if absent

volumes:
  pgdata:
```

---

## docker-compose.e2e.yml

Identical to `docker-compose.yml` except:
- Different port: `8081:8080`
- Different named network + volumes (isolated from dev stack)
- Runs migrations automatically on startup (via `command: sh -c "migrate ... && /server"`)
- DB: `workflow_e2e`

---

## Graceful Shutdown (`backend/cmd/server/main.go`)

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        log.Error("server error", "err", err)
        os.Exit(1)
    }
}()

go worker.Run(ctx)

<-ctx.Done()
log.Info("shutdown signal received")

shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(shutCtx); err != nil {
    log.Error("HTTP shutdown error", "err", err)
}
worker.Stop(shutCtx)
pool.Close()
log.Info("shutdown complete")
```

---

## Makefile (final targets)

```makefile
.PHONY: build run test test-integration docker-up docker-down \
        migrate-up migrate-up-test migrate-down \
        frontend-dev frontend-build \
        stack-up stack-down smoke e2e

BINARY      := ./bin/server
DB_URL      := postgres://workflow:workflow@localhost:5432/workflow?sslmode=disable
TEST_DB_URL := postgres://workflow:workflow@localhost:5432/workflow_test?sslmode=disable

build:
	go build -o $(BINARY) ./backend/cmd/server

run: build
	DATABASE_URL=$(DB_URL) $(BINARY)

test:
	go test ./backend/...

test-integration:
	DATABASE_URL=$(TEST_DB_URL) go test ./backend/tests/integration/... -v -count=1

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

migrate-up:
	migrate -path backend/migrations -database "$(DB_URL)" up

migrate-up-test:
	migrate -path backend/migrations -database "$(TEST_DB_URL)" up

migrate-down:
	migrate -path backend/migrations -database "$(DB_URL)" down 1

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm test

stack-up:
	docker compose up --build -d

stack-down:
	docker compose down -v

smoke: stack-up
	./tests/smoke/smoke.sh

e2e:
	docker compose -f docker-compose.e2e.yml up --build -d
	cd tests/e2e && npm ci && npx playwright test
	docker compose -f docker-compose.e2e.yml down -v
```

---

## Smoke Test (`tests/smoke/smoke.sh`)

curl-based, tests golden path against running stack (assumes `http://localhost:8080`):

```bash
#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080/api/v1"

echo "=== Health ==="
curl -sf "$BASE/../healthz" | grep -q '"ok"' && echo "OK"

echo "=== List defs ==="
DEFS=$(curl -sf "$BASE/workflow-defs")
echo "$DEFS" | python3 -m json.tool > /dev/null && echo "OK"

echo "=== Get first def ==="
DEF_ID=$(echo "$DEFS" | python3 -c "import sys,json; print(json.load(sys.stdin)['items'][0]['id'])")

echo "=== Start instance ==="
INST=$(curl -sf -X POST "$BASE/workflow-instances" \
  -H "Content-Type: application/json" \
  -d "{\"workflow_def_id\":\"$DEF_ID\",\"created_by\":\"smoke-test\"}")
INST_ID=$(echo "$INST" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo "Instance: $INST_ID"

echo "=== Get instance ==="
curl -sf "$BASE/workflow-instances/$INST_ID" | python3 -m json.tool > /dev/null && echo "OK"

echo "=== Get tasks ==="
TASKS=$(curl -sf "$BASE/workflow-instances/$INST_ID/tasks")
TASK_ID=$(echo "$TASKS" | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['id'])")

echo "=== Complete task ==="
curl -sf -X POST "$BASE/task-instances/$TASK_ID/complete" \
  -H "Content-Type: application/json" \
  -d '{"status":"done","completed_by":"smoke-test"}' | python3 -m json.tool > /dev/null && echo "OK"

echo "=== Smoke tests passed ==="
```

---

## Playwright E2e Tests (`tests/e2e/`)

### `playwright.config.ts`
```ts
export default defineConfig({
  testDir: './specs',
  use: {
    baseURL: 'http://localhost:3001',    // frontend via docker-compose.e2e.yml
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  retries: 1,
})
```

### Fixture (`fixtures/leave-approval.json`)
Pre-loaded workflow def used by all e2e specs.

### DB Reset Helper
Each spec calls a helper that hits `POST /api/v1/admin/test/reset-db` (test-only endpoint, registered only when `APP_ENV=test`) OR runs direct psql truncate via `execSync`.

### Spec Coverage

| File | Scenario |
|------|---------|
| `golden-path.spec.ts` | Open app → see def → start instance → complete all tasks in order → instance shows completed |
| `cancel.spec.ts` | Start instance → cancel → instance shows cancelled → tasks show cancelled |
| `rejection.spec.ts` | Start instance → reject task → instance cancelled |
| `decision-state.spec.ts` | Multi-edge (user-picks) → next_state dropdown present → pick one → correct state entered |
| `task-bound.spec.ts` | Task-bound state → complete bound task → sibling shown as cancelled |
| `admin-reload.spec.ts` | Navigate to /admin/reload → click Reload → result shows loaded + warnings |
| `list-filter.spec.ts` | Start multiple instances → filter by status → correct subset shown |

---

## README.md (root)

Sections:
1. **What** — one paragraph, problem statement
2. **Quickstart** — `docker compose up --build` → open `http://localhost:8080`
3. **Define a Workflow** — sample JSON walk-through, where to put files
4. **Dev Setup** — local Go + Vite + Postgres step-by-step
5. **API Examples** — start instance, complete task (curl snippets)
6. **Webhooks** — `webhooks.json` format, payload, signature verification
7. **Environment Variables** — table (from Architecture.md)
8. **Testing** — `make test`, `make test-integration`, `make e2e`
9. **Architecture** — link to `spec/Architecture.md`
10. **Status** — link to `spec/Currentstatus.md`

---

## Security Checklist (final pass)

- [ ] Error responses never expose stack traces, SQL, or internal paths
- [ ] Webhook secret not logged
- [ ] `webhooks.json` in `.gitignore`
- [ ] Input validation on all handler entry points (UUIDs, enums, required fields)
- [ ] Pagination max limit enforced (200)
- [ ] `X-Content-Type-Options: nosniff` header on API responses
- [ ] Panic recovery returns generic 500 (no panic message leaked)
- [ ] `SELECT FOR UPDATE` confirms concurrency protection end-to-end

---

## Exit Criteria
- [ ] `make stack-up && make smoke` passes
- [ ] `make e2e` — all 7 Playwright specs green
- [ ] `docker compose up --build` — app loads in browser at `http://localhost:8080`
- [ ] SIGTERM triggers clean shutdown (manual verify: `docker stop`)
- [ ] Security checklist all checked
- [ ] README quickstart works from a clean clone
