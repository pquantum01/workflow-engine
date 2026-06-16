package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type workflowDefRepo struct {
	db DB
}

func NewWorkflowDefRepo(db DB) WorkflowDefRepo {
	return &workflowDefRepo{db: db}
}

func (r *workflowDefRepo) WithTx(tx DB) WorkflowDefRepo {
	return &workflowDefRepo{db: tx}
}

func (r *workflowDefRepo) Create(ctx context.Context, def *model.WorkflowDef) error {
	defJSON, err := json.Marshal(def.Definition)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO workflow_def (id, name, version, description, definition, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		def.ID, def.Name, def.Version, def.Description, defJSON, def.CreatedAt,
	)
	return err
}

func (r *workflowDefRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowDef, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, version, description, definition, created_at, deleted_at
		FROM workflow_def WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanWorkflowDef(row)
}

func (r *workflowDefRepo) GetByNameVersion(ctx context.Context, name string, version int) (*model.WorkflowDef, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, version, description, definition, created_at, deleted_at
		FROM workflow_def WHERE name = $1 AND version = $2 AND deleted_at IS NULL`, name, version)
	return scanWorkflowDef(row)
}

func (r *workflowDefRepo) List(ctx context.Context, includeDeleted bool) ([]*model.WorkflowDef, error) {
	q := `SELECT id, name, version, description, definition, created_at, deleted_at FROM workflow_def`
	if !includeDeleted {
		q += ` WHERE deleted_at IS NULL`
	}
	q += ` ORDER BY name, version`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var defs []*model.WorkflowDef
	for rows.Next() {
		def, err := scanWorkflowDefRow(rows)
		if err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}
	return defs, rows.Err()
}

func (r *workflowDefRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workflow_def SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now(), id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanWorkflowDef(row rowScanner) (*model.WorkflowDef, error) {
	var def model.WorkflowDef
	var defJSON []byte
	var deletedAt sql.NullTime
	err := row.Scan(&def.ID, &def.Name, &def.Version, &def.Description, &defJSON, &def.CreatedAt, &deletedAt)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if deletedAt.Valid {
		def.DeletedAt = &deletedAt.Time
	}
	if err := json.Unmarshal(defJSON, &def.Definition); err != nil {
		return nil, err
	}
	return &def, nil
}

func scanWorkflowDefRow(rows *sql.Rows) (*model.WorkflowDef, error) {
	var def model.WorkflowDef
	var defJSON []byte
	var deletedAt sql.NullTime
	err := rows.Scan(&def.ID, &def.Name, &def.Version, &def.Description, &defJSON, &def.CreatedAt, &deletedAt)
	if err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		def.DeletedAt = &deletedAt.Time
	}
	if err := json.Unmarshal(defJSON, &def.Definition); err != nil {
		return nil, err
	}
	return &def, nil
}
