//go:build integration

package integration

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://workflow:workflow@localhost:5433/workflow_engine_test?sslmode=disable&search_path=workflow_engine"
	}

	var err error
	testDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalf("failed to ping test database: %v", err)
	}

	code := m.Run()

	testDB.Close()
	os.Exit(code)
}

func cleanupTables(t *testing.T) {
	t.Helper()
	tables := []string{
		"state_transitions",
		"task_instances",
		"workflow_instances",
		"process_instances",
		"task",
		"workflow",
		"process",
	}
	for _, table := range tables {
		if _, err := testDB.Exec("DELETE FROM " + table); err != nil {
			t.Fatalf("failed to clean table %s: %v", table, err)
		}
	}
}
