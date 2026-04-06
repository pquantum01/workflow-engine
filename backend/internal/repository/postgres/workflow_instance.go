package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type WorkflowInstanceRepo struct {
	db *sql.DB
}

func NewWorkflowInstanceRepo(db *sql.DB) *WorkflowInstanceRepo {
	return &WorkflowInstanceRepo{db: db}
}

func (r *WorkflowInstanceRepo) Create(ctx context.Context, wi *model.WorkflowInstance) error {
	query := `INSERT INTO workflow_instances (id, workflow_id, process_instance_id, current_state, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		wi.ID, wi.WorkflowID, wi.ProcessInstanceID, wi.CurrentState, wi.Status, wi.CreatedAt, wi.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *WorkflowInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	query := `SELECT id, workflow_id, process_instance_id, current_state, status, created_at, updated_at
		FROM workflow_instances WHERE id = $1`
	wi := &model.WorkflowInstance{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wi.ID, &wi.WorkflowID, &wi.ProcessInstanceID, &wi.CurrentState, &wi.Status, &wi.CreatedAt, &wi.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return wi, nil
}

func (r *WorkflowInstanceRepo) ListByProcessInstanceID(ctx context.Context, processInstanceID uuid.UUID) ([]*model.WorkflowInstance, error) {
	query := `SELECT id, workflow_id, process_instance_id, current_state, status, created_at, updated_at
		FROM workflow_instances WHERE process_instance_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, processInstanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.WorkflowInstance
	for rows.Next() {
		wi := &model.WorkflowInstance{}
		if err := rows.Scan(&wi.ID, &wi.WorkflowID, &wi.ProcessInstanceID, &wi.CurrentState, &wi.Status, &wi.CreatedAt, &wi.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, wi)
	}
	return results, rows.Err()
}

func (r *WorkflowInstanceRepo) Update(ctx context.Context, wi *model.WorkflowInstance) error {
	query := `UPDATE workflow_instances SET current_state = $1, status = $2, updated_at = $3
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, wi.CurrentState, wi.Status, wi.UpdatedAt, wi.ID)
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
