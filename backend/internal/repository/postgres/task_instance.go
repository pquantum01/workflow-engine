package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type TaskInstanceRepo struct {
	db *sql.DB
}

func NewTaskInstanceRepo(db *sql.DB) *TaskInstanceRepo {
	return &TaskInstanceRepo{db: db}
}

func (r *TaskInstanceRepo) Create(ctx context.Context, ti *model.TaskInstance) error {
	query := `INSERT INTO task_instances (id, task_id, workflow_instance_id, status, assigned_to, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		ti.ID, ti.TaskID, ti.WorkflowInstanceID, ti.Status, ti.AssignedTo, ti.CompletedAt, ti.CreatedAt, ti.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *TaskInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.TaskInstance, error) {
	query := `SELECT id, task_id, workflow_instance_id, status, assigned_to, completed_at, created_at, updated_at
		FROM task_instances WHERE id = $1`
	ti := &model.TaskInstance{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ti.ID, &ti.TaskID, &ti.WorkflowInstanceID, &ti.Status, &ti.AssignedTo, &ti.CompletedAt, &ti.CreatedAt, &ti.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return ti, nil
}

func (r *TaskInstanceRepo) ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.TaskInstance, error) {
	query := `SELECT id, task_id, workflow_instance_id, status, assigned_to, completed_at, created_at, updated_at
		FROM task_instances WHERE workflow_instance_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, workflowInstanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.TaskInstance
	for rows.Next() {
		ti := &model.TaskInstance{}
		if err := rows.Scan(&ti.ID, &ti.TaskID, &ti.WorkflowInstanceID, &ti.Status, &ti.AssignedTo, &ti.CompletedAt, &ti.CreatedAt, &ti.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, ti)
	}
	return results, rows.Err()
}

func (r *TaskInstanceRepo) Update(ctx context.Context, ti *model.TaskInstance) error {
	query := `UPDATE task_instances SET status = $1, assigned_to = $2, completed_at = $3, updated_at = $4
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, ti.Status, ti.AssignedTo, ti.CompletedAt, ti.UpdatedAt, ti.ID)
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
