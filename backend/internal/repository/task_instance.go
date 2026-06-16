package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type taskInstanceRepo struct {
	db DB
}

func NewTaskInstanceRepo(db DB) TaskInstanceRepo {
	return &taskInstanceRepo{db: db}
}

func (r *taskInstanceRepo) WithTx(tx DB) TaskInstanceRepo {
	return &taskInstanceRepo{db: tx}
}

func (r *taskInstanceRepo) BulkCreate(ctx context.Context, tasks []*model.TaskInstance) error {
	if len(tasks) == 0 {
		return nil
	}
	valueStrings := make([]string, 0, len(tasks))
	args := make([]any, 0, len(tasks)*9)
	for i, t := range tasks {
		base := i * 9
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9,
		))
		args = append(args,
			t.ID, t.WorkflowInstanceID, t.StateID, t.TaskKey,
			t.TaskLabel, t.TaskType, string(t.Status), t.CreatedBy, t.CreatedAt,
		)
	}
	q := fmt.Sprintf(`INSERT INTO task_instance
		(id, workflow_instance_id, state_id, task_key, task_label, task_type, status, created_by, created_at)
		VALUES %s`, strings.Join(valueStrings, ","))
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

func (r *taskInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workflow_instance_id, state_id, task_key, task_label, task_type,
		       status, outcome, data, created_by, created_at, completed_at, deleted_at
		FROM task_instance WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanTask(row)
}

func (r *taskInstanceRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error) {
	q := `SELECT id, workflow_instance_id, state_id, task_key, task_label, task_type,
		       status, outcome, data, created_by, created_at, completed_at, deleted_at
		FROM task_instance WHERE workflow_instance_id = $1 AND deleted_at IS NULL`
	args := []any{instanceID}

	if len(statusFilter) > 0 {
		placeholders := make([]string, len(statusFilter))
		for i, s := range statusFilter {
			args = append(args, string(s))
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		q += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}
	q += " ORDER BY created_at"

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTaskRows(rows)
}

func (r *taskInstanceRepo) ListByInstanceAndState(ctx context.Context, instanceID uuid.UUID, stateID string) ([]*model.TaskInstance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workflow_instance_id, state_id, task_key, task_label, task_type,
		       status, outcome, data, created_by, created_at, completed_at, deleted_at
		FROM task_instance
		WHERE workflow_instance_id = $1 AND state_id = $2 AND deleted_at IS NULL
		ORDER BY created_at`,
		instanceID, stateID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTaskRows(rows)
}

func (r *taskInstanceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TaskStatus, completedAt *time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE task_instance SET status = $1, completed_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
		string(status), completedAt, id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (r *taskInstanceRepo) UpdateOutcome(ctx context.Context, id uuid.UUID, outcome string, data []byte) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE task_instance SET outcome = $1, data = $2 WHERE id = $3 AND deleted_at IS NULL`,
		outcome, data, id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (r *taskInstanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE task_instance SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now(), id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func scanTask(row rowScanner) (*model.TaskInstance, error) {
	var t model.TaskInstance
	var status string
	var outcome sql.NullString
	var completedAt, deletedAt sql.NullTime
	err := row.Scan(
		&t.ID, &t.WorkflowInstanceID, &t.StateID, &t.TaskKey, &t.TaskLabel, &t.TaskType,
		&status, &outcome, &t.Data, &t.CreatedBy, &t.CreatedAt, &completedAt, &deletedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t.Status = model.TaskStatus(status)
	if outcome.Valid {
		t.Outcome = &outcome.String
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if deletedAt.Valid {
		t.DeletedAt = &deletedAt.Time
	}
	return &t, nil
}

func scanTaskRows(rows *sql.Rows) ([]*model.TaskInstance, error) {
	var tasks []*model.TaskInstance
	for rows.Next() {
		var t model.TaskInstance
		var status string
		var outcome sql.NullString
		var completedAt, deletedAt sql.NullTime
		err := rows.Scan(
			&t.ID, &t.WorkflowInstanceID, &t.StateID, &t.TaskKey, &t.TaskLabel, &t.TaskType,
			&status, &outcome, &t.Data, &t.CreatedBy, &t.CreatedAt, &completedAt, &deletedAt,
		)
		if err != nil {
			return nil, err
		}
		t.Status = model.TaskStatus(status)
		if outcome.Valid {
			t.Outcome = &outcome.String
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if deletedAt.Valid {
			t.DeletedAt = &deletedAt.Time
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}
