package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type workflowInstanceRepo struct {
	db DB
}

func NewWorkflowInstanceRepo(db DB) WorkflowInstanceRepo {
	return &workflowInstanceRepo{db: db}
}

func (r *workflowInstanceRepo) WithTx(tx DB) WorkflowInstanceRepo {
	return &workflowInstanceRepo{db: tx}
}

func (r *workflowInstanceRepo) Create(ctx context.Context, inst *model.WorkflowInstance) error {
	ctxJSON, err := json.Marshal(inst.Context)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO workflow_instance (id, workflow_def_id, current_state_id, status, context, created_by, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		inst.ID, inst.WorkflowDefID, inst.CurrentStateID, string(inst.Status), ctxJSON, inst.CreatedBy, inst.StartedAt,
	)
	return err
}

func (r *workflowInstanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workflow_def_id, current_state_id, status, context, created_by, started_at, completed_at, deleted_at
		FROM workflow_instance WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanInstance(row)
}

func (r *workflowInstanceRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.WorkflowInstance, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, workflow_def_id, current_state_id, status, context, created_by, started_at, completed_at, deleted_at
		FROM workflow_instance WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, id)
	return scanInstance(row)
}

func (r *workflowInstanceRepo) List(ctx context.Context, filter InstanceFilter) ([]*model.WorkflowInstance, int, error) {
	where := "deleted_at IS NULL"
	args := []any{}
	argN := 1

	if filter.DefID != nil {
		where += fmt.Sprintf(" AND workflow_def_id = $%d", argN)
		args = append(args, *filter.DefID)
		argN++
	}
	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, string(*filter.Status))
		argN++
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workflow_instance WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	args = append(args, limit, filter.Offset)
	q := fmt.Sprintf(`SELECT id, workflow_def_id, current_state_id, status, context, created_by, started_at, completed_at, deleted_at
		FROM workflow_instance WHERE %s ORDER BY started_at DESC LIMIT $%d OFFSET $%d`,
		where, argN, argN+1)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var insts []*model.WorkflowInstance
	for rows.Next() {
		inst, err := scanInstanceRow(rows)
		if err != nil {
			return nil, 0, err
		}
		insts = append(insts, inst)
	}
	return insts, total, rows.Err()
}

func (r *workflowInstanceRepo) UpdateState(ctx context.Context, id uuid.UUID, newStateID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workflow_instance SET current_state_id = $1 WHERE id = $2 AND deleted_at IS NULL`,
		newStateID, id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (r *workflowInstanceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.InstanceStatus, completedAt *time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workflow_instance SET status = $1, completed_at = $2 WHERE id = $3 AND deleted_at IS NULL`,
		string(status), completedAt, id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func (r *workflowInstanceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workflow_instance SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		time.Now(), id,
	)
	if err != nil {
		return err
	}
	return requireOneRow(res)
}

func scanInstance(row rowScanner) (*model.WorkflowInstance, error) {
	var inst model.WorkflowInstance
	var ctxJSON []byte
	var status string
	var completedAt, deletedAt sql.NullTime
	err := row.Scan(
		&inst.ID, &inst.WorkflowDefID, &inst.CurrentStateID,
		&status, &ctxJSON, &inst.CreatedBy, &inst.StartedAt, &completedAt, &deletedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	inst.Status = model.InstanceStatus(status)
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}
	if err := json.Unmarshal(ctxJSON, &inst.Context); err != nil {
		inst.Context = map[string]any{}
	}
	return &inst, nil
}

func scanInstanceRow(rows *sql.Rows) (*model.WorkflowInstance, error) {
	var inst model.WorkflowInstance
	var ctxJSON []byte
	var status string
	var completedAt, deletedAt sql.NullTime
	err := rows.Scan(
		&inst.ID, &inst.WorkflowDefID, &inst.CurrentStateID,
		&status, &ctxJSON, &inst.CreatedBy, &inst.StartedAt, &completedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}
	inst.Status = model.InstanceStatus(status)
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}
	if deletedAt.Valid {
		inst.DeletedAt = &deletedAt.Time
	}
	if err := json.Unmarshal(ctxJSON, &inst.Context); err != nil {
		inst.Context = map[string]any{}
	}
	return &inst, nil
}

func requireOneRow(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
