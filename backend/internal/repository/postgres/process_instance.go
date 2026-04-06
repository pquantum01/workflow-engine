package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type ProcessInstanceRepo struct {
	db *sql.DB
}

func NewProcessInstanceRepo(db *sql.DB) *ProcessInstanceRepo {
	return &ProcessInstanceRepo{db: db}
}

func (r *ProcessInstanceRepo) Create(ctx context.Context, pi *model.ProcessInstance) error {
	query := `INSERT INTO process_instances (id, process_id, parent_id, status, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		pi.ID, pi.ProcessID, pi.ParentID, pi.Status, pi.StartedAt, pi.CompletedAt, pi.CreatedAt, pi.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *ProcessInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error) {
	query := `SELECT id, process_id, parent_id, status, started_at, completed_at, created_at, updated_at
		FROM process_instances WHERE id = $1`
	pi := &model.ProcessInstance{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pi.ID, &pi.ProcessID, &pi.ParentID, &pi.Status, &pi.StartedAt, &pi.CompletedAt, &pi.CreatedAt, &pi.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return pi, nil
}

func (r *ProcessInstanceRepo) List(ctx context.Context, offset, limit int) ([]*model.ProcessInstance, error) {
	query := `SELECT id, process_id, parent_id, status, started_at, completed_at, created_at, updated_at
		FROM process_instances ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.ProcessInstance
	for rows.Next() {
		pi := &model.ProcessInstance{}
		if err := rows.Scan(&pi.ID, &pi.ProcessID, &pi.ParentID, &pi.Status, &pi.StartedAt, &pi.CompletedAt, &pi.CreatedAt, &pi.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, pi)
	}
	return results, rows.Err()
}

func (r *ProcessInstanceRepo) ListByParentID(ctx context.Context, parentID uuid.UUID) ([]*model.ProcessInstance, error) {
	query := `SELECT id, process_id, parent_id, status, started_at, completed_at, created_at, updated_at
		FROM process_instances WHERE parent_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*model.ProcessInstance
	for rows.Next() {
		pi := &model.ProcessInstance{}
		if err := rows.Scan(&pi.ID, &pi.ProcessID, &pi.ParentID, &pi.Status, &pi.StartedAt, &pi.CompletedAt, &pi.CreatedAt, &pi.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, pi)
	}
	return results, rows.Err()
}

func (r *ProcessInstanceRepo) Update(ctx context.Context, pi *model.ProcessInstance) error {
	query := `UPDATE process_instances SET status = $1, completed_at = $2, updated_at = $3
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, pi.Status, pi.CompletedAt, pi.UpdatedAt, pi.ID)
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
