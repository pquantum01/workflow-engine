.PHONY: build test test-integration migrate-up migrate-down docker-up docker-down lint run

# Backend
build:
	go build -o bin/server ./backend/cmd/server

run: build
	./bin/server

test:
	go test ./backend/internal/...

test-integration:
	go test -tags=integration ./backend/tests/integration/...

lint:
	go vet ./backend/...

# Database migrations
MIGRATE=go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	$(MIGRATE) -path backend/migrations -database "postgres://workflow:workflow@localhost:5432/workflow_engine?sslmode=disable" up

migrate-down:
	$(MIGRATE) -path backend/migrations -database "postgres://workflow:workflow@localhost:5432/workflow_engine?sslmode=disable" down 1

migrate-up-test:
	$(MIGRATE) -path backend/migrations -database "postgres://workflow:workflow@localhost:5433/workflow_engine_test?sslmode=disable" up

migrate-down-test:
	$(MIGRATE) -path backend/migrations -database "postgres://workflow:workflow@localhost:5433/workflow_engine_test?sslmode=disable" down 1

# Docker
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
