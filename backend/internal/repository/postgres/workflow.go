package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type WorkflowRepo struct {
	db *sql.DB
}

func NewWorkflowRepo(db *sql.DB) *WorkflowRepo {
	return &WorkflowRepo{db: db}
}

func (r *WorkflowRepo) Create(ctx context.Context, w *model.Workflow) error {
	smJSON, err := json.Marshal(w.StateMachine)
	if err != nil {
		return err
	}
	query := `INSERT INTO workflow (id, process_id, name, description, state_machine, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = r.db.ExecContext(ctx, query,
		w.ID, w.ProcessID, w.Name, w.Description, smJSON, w.SortOrder, w.CreatedAt, w.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *WorkflowRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Workflow, error) {
	query := `SELECT id, process_id, name, description, state_machine, sort_order, created_at, updated_at
		FROM workflow WHERE id = $1`
	w := &model.Workflow{}
	var smJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.ProcessID, &w.Name, &w.Description, &smJSON, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(smJSON, &w.StateMachine); err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WorkflowRepo) ListByProcessID(ctx context.Context, processID uuid.UUID) ([]*model.Workflow, error) {
	query := `SELECT id, process_id, name, description, state_machine, sort_order, created_at, updated_at
		FROM workflow WHERE process_id = $1 ORDER BY sort_order`
	rows, err := r.db.QueryContext(ctx, query, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.Workflow, 0)
	for rows.Next() {
		w := &model.Workflow{}
		var smJSON []byte
		if err := rows.Scan(&w.ID, &w.ProcessID, &w.Name, &w.Description, &smJSON, &w.SortOrder, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(smJSON, &w.StateMachine); err != nil {
			return nil, err
		}
		results = append(results, w)
	}
	return results, rows.Err()
}

func (r *WorkflowRepo) Update(ctx context.Context, w *model.Workflow) error {
	smJSON, err := json.Marshal(w.StateMachine)
	if err != nil {
		return err
	}
	query := `UPDATE workflow SET process_id = $1, name = $2, description = $3, state_machine = $4, sort_order = $5, updated_at = $6
		WHERE id = $7`
	result, err := r.db.ExecContext(ctx, query, w.ProcessID, w.Name, w.Description, smJSON, w.SortOrder, w.UpdatedAt, w.ID)
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

func (r *WorkflowRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workflow WHERE id = $1`
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
