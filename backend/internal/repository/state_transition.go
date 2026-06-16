package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type stateTransitionRepo struct {
	db DB
}

func NewStateTransitionRepo(db DB) StateTransitionRepo {
	return &stateTransitionRepo{db: db}
}

func (r *stateTransitionRepo) WithTx(tx DB) StateTransitionRepo {
	return &stateTransitionRepo{db: tx}
}

func (r *stateTransitionRepo) Append(ctx context.Context, t *model.StateTransition) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO state_transition (id, workflow_instance_id, from_state, to_state, transitioned_at)
		VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.WorkflowInstanceID, t.FromState, t.ToState, t.TransitionedAt,
	)
	return err
}

func (r *stateTransitionRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]*model.StateTransition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, workflow_instance_id, from_state, to_state, transitioned_at
		FROM state_transition WHERE workflow_instance_id = $1 ORDER BY transitioned_at`,
		instanceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transitions []*model.StateTransition
	for rows.Next() {
		var t model.StateTransition
		if err := rows.Scan(&t.ID, &t.WorkflowInstanceID, &t.FromState, &t.ToState, &t.TransitionedAt); err != nil {
			return nil, err
		}
		transitions = append(transitions, &t)
	}
	return transitions, rows.Err()
}
