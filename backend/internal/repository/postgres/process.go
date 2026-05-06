package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
	"workflow-engine/backend/internal/repository"
)

type ProcessRepo struct {
	db *sql.DB
}

func NewProcessRepo(db *sql.DB) *ProcessRepo {
	return &ProcessRepo{db: db}
}

func (r *ProcessRepo) Create(ctx context.Context, p *model.Process) error {
	query := `INSERT INTO process (id, name, description, version, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Version, p.Status, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *ProcessRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Process, error) {
	query := `SELECT id, name, description, version, status, created_at, updated_at
		FROM process WHERE id = $1`
	p := &model.Process{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Version, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProcessRepo) List(ctx context.Context, offset, limit int) ([]*model.Process, error) {
	query := `SELECT id, name, description, version, status, created_at, updated_at
		FROM process ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.Process, 0)
	for rows.Next() {
		p := &model.Process{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Version, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

func (r *ProcessRepo) Update(ctx context.Context, p *model.Process) error {
	query := `UPDATE process SET name = $1, description = $2, version = $3, status = $4, updated_at = $5
		WHERE id = $6`
	result, err := r.db.ExecContext(ctx, query, p.Name, p.Description, p.Version, p.Status, p.UpdatedAt, p.ID)
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

func (r *ProcessRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM process WHERE id = $1`
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
