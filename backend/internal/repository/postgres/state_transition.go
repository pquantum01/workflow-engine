package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"workflow-engine/backend/internal/model"
)

type StateTransitionRepo struct {
	db *sql.DB
}

func NewStateTransitionRepo(db *sql.DB) *StateTransitionRepo {
	return &StateTransitionRepo{db: db}
}

func (r *StateTransitionRepo) Create(ctx context.Context, st *model.StateTransition) error {
	query := `INSERT INTO state_transitions (id, workflow_instance_id, from_state, to_state, transition_name, triggered_by, trigger_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		st.ID, st.WorkflowInstanceID, st.FromState, st.ToState, st.TransitionName, st.TriggeredBy, st.TriggerType, st.CreatedAt)
	if err != nil {
		return wrapDBError(err)
	}
	return nil
}

func (r *StateTransitionRepo) ListByWorkflowInstanceID(ctx context.Context, workflowInstanceID uuid.UUID) ([]*model.StateTransition, error) {
	query := `SELECT id, workflow_instance_id, from_state, to_state, transition_name, triggered_by, trigger_type, created_at
		FROM state_transitions WHERE workflow_instance_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, workflowInstanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*model.StateTransition, 0)
	for rows.Next() {
		st := &model.StateTransition{}
		if err := rows.Scan(&st.ID, &st.WorkflowInstanceID, &st.FromState, &st.ToState, &st.TransitionName, &st.TriggeredBy, &st.TriggerType, &st.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, st)
	}
	return results, rows.Err()
}
