.PHONY: build run test test-integration docker-up docker-down \
        migrate-up migrate-up-test migrate-down \
        frontend-dev frontend-build frontend-test \
        stack-up stack-down smoke e2e

BINARY      := ./bin/server
DB_URL      := postgres://workflow:workflow@localhost:5432/workflow?sslmode=disable
TEST_DB_URL := postgres://workflow:workflow@localhost:5433/workflow_test?sslmode=disable

build:
	mkdir -p bin
	cd backend && go build -o ../bin/server ./cmd/server

run: build
	DATABASE_URL=$(DB_URL) \
	LOG_LEVEL=debug \
	LOG_FORMAT=text \
	WORKFLOWS_DIR=./workflows \
	WEBHOOK_CONFIG=./webhooks.json \
	$(BINARY)

test:
	cd backend && go test ./...

test-integration:
	cd backend && DATABASE_URL=$(TEST_DB_URL) go test ./tests/integration/... -v -count=1

docker-up:
	docker-compose up -d postgres postgres_test

docker-down:
	docker-compose down

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
	docker-compose -f docker-compose.yml up --build -d

stack-down:
	docker-compose -f docker-compose.yml down -v

smoke: stack-up
	./tests/smoke/smoke.sh

e2e:
	docker compose -f docker-compose.e2e.yml up --build -d
	cd tests/e2e && npm ci && npx playwright test
	docker compose -f docker-compose.e2e.yml down -v
