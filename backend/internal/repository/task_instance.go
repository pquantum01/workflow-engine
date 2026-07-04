package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
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
	args := make([]any, 0, len(tasks)*12)
	for i, t := range tasks {
		base := i * 12
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12,
		))
		priority := t.Priority
		if priority == "" {
			priority = "medium"
		}
		args = append(args,
			t.ID, t.WorkflowInstanceID, t.StateID, t.TaskKey,
			t.TaskLabel, t.TaskType, string(t.Status), t.CreatedBy, t.CreatedAt,
			priority, t.IsAdhoc, t.ExpectedCompletionDate,
		)
	}
	q := fmt.Sprintf(`INSERT INTO task_instance
		(id, workflow_instance_id, state_id, task_key, task_label, task_type, status, created_by, created_at,
		 priority, is_adhoc, expected_completion_date)
		VALUES %s`, strings.Join(valueStrings, ","))
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

func (r *taskInstanceRepo) Create(ctx context.Context, task *model.TaskInstance) error {
	priority := task.Priority
	if priority == "" {
		priority = "medium"
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_instance
		  (id, workflow_instance_id, state_id, task_key, task_label, task_type,
		   status, created_by, created_at, priority, is_adhoc,
		   expected_completion_date, expected_completion_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		task.ID, task.WorkflowInstanceID, task.StateID, task.TaskKey, task.TaskLabel,
		task.TaskType, string(task.Status), task.CreatedBy, task.CreatedAt,
		priority, task.IsAdhoc,
		task.ExpectedCompletionDate, task.ExpectedCompletionTime,
	)
	return err
}

func (r *taskInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workflow_instance_id, state_id, task_key, task_label, task_type,
		       status, outcome, data, created_by, created_at, completed_at, deleted_at,
		       priority, is_adhoc, expected_completion_date::TEXT, expected_completion_time::TEXT
		FROM task_instance WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanTask(row)
}

func (r *taskInstanceRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID, statusFilter []model.TaskStatus) ([]*model.TaskInstance, error) {
	q := `SELECT id, workflow_instance_id, state_id, task_key, task_label, task_type,
		       status, outcome, data, created_by, created_at, completed_at, deleted_at,
		       priority, is_adhoc, expected_completion_date::TEXT, expected_completion_time::TEXT
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
		       status, outcome, data, created_by, created_at, completed_at, deleted_at,
		       priority, is_adhoc, expected_completion_date::TEXT, expected_completion_time::TEXT
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

func (r *taskInstanceRepo) UpdateFields(ctx context.Context, id uuid.UUID, patch TaskInstancePatch) error {
	var setClauses []string
	var args []any
	argIdx := 1

	if patch.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", argIdx))
		args = append(args, *patch.Priority)
		argIdx++
	}
	if patch.ExpectedCompletionDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("expected_completion_date = $%d", argIdx))
		if *patch.ExpectedCompletionDate == "" {
			args = append(args, nil)
		} else {
			args = append(args, *patch.ExpectedCompletionDate)
		}
		argIdx++
	}
	if patch.ExpectedCompletionTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("expected_completion_time = $%d", argIdx))
		if *patch.ExpectedCompletionTime == "" {
			args = append(args, nil)
		} else {
			args = append(args, *patch.ExpectedCompletionTime)
		}
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	args = append(args, id)
	q := fmt.Sprintf("UPDATE task_instance SET %s WHERE id = $%d AND deleted_at IS NULL",
		strings.Join(setClauses, ", "), argIdx)
	res, err := r.db.ExecContext(ctx, q, args...)
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
	var expDate, expTime sql.NullString
	err := row.Scan(
		&t.ID, &t.WorkflowInstanceID, &t.StateID, &t.TaskKey, &t.TaskLabel, &t.TaskType,
		&status, &outcome, &t.Data, &t.CreatedBy, &t.CreatedAt, &completedAt, &deletedAt,
		&t.Priority, &t.IsAdhoc, &expDate, &expTime,
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
	if expDate.Valid {
		t.ExpectedCompletionDate = &expDate.String
	}
	if expTime.Valid {
		t.ExpectedCompletionTime = &expTime.String
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
		var expDate, expTime sql.NullString
		err := rows.Scan(
			&t.ID, &t.WorkflowInstanceID, &t.StateID, &t.TaskKey, &t.TaskLabel, &t.TaskType,
			&status, &outcome, &t.Data, &t.CreatedBy, &t.CreatedAt, &completedAt, &deletedAt,
			&t.Priority, &t.IsAdhoc, &expDate, &expTime,
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
		if expDate.Valid {
			t.ExpectedCompletionDate = &expDate.String
		}
		if expTime.Valid {
			t.ExpectedCompletionTime = &expTime.String
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}
