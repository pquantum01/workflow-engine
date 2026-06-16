package model

import (
	"time"

	"github.com/google/uuid"
)

type StateTransition struct {
	ID                 uuid.UUID
	WorkflowInstanceID uuid.UUID
	FromState          string
	ToState            string
	TransitionedAt     time.Time
}
