package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *model.Task) error {
	query := `INSERT INTO task (id, workflow_id, name, description, task_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		t.ID, t.WorkflowID, t.Name, t.Description, t.TaskType, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	query := `SELECT id, workflow_id, name, description, task_type, created_at, updated_at
		FROM task WHERE id = $1`
	t := &model.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.WorkflowID, &t.Name, &t.Description, &t.TaskType, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepo) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*model.Task, error) {
	query := `SELECT id, workflow_id, name, description, task_type, created_at, updated_at
		FROM task WHERE workflow_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.Task, 0)
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.WorkflowID, &t.Name, &t.Description, &t.TaskType, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

func (r *TaskRepo) Update(ctx context.Context, t *model.Task) error {
	query := `UPDATE task SET workflow_id = $1, name = $2, description = $3, task_type = $4, updated_at = $5
		WHERE id = $6`
	result, err := r.db.ExecContext(ctx, query, t.WorkflowID, t.Name, t.Description, t.TaskType, t.UpdatedAt, t.ID)
	if err != nil {
		return wrapDBError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM task WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return wrapDBError(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrNotFound
	}
	return nil
}
