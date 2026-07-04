package integration_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://workflow:workflow@localhost:5433/workflow_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("failed to open test DB: " + err.Error())
	}
	if err := db.PingContext(context.Background()); err != nil {
		panic("test DB ping failed: " + err.Error())
	}
	testDB = db

	code := m.Run()
	db.Close()
	os.Exit(code)
}

func truncateTables(t *testing.T) {
	t.Helper()
	_, err := testDB.ExecContext(context.Background(), `
		TRUNCATE TABLE
			webhook_outbox,
			task_attachment,
			task_comment,
			state_transition,
			task_instance,
			workflow_instance,
			workflow_def
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}
