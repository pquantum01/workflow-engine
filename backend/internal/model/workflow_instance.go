package model

import (
	"time"

	"github.com/google/uuid"
)

type InstanceStatus string

const (
	InstanceStatusActive    InstanceStatus = "active"
	InstanceStatusCompleted InstanceStatus = "completed"
	InstanceStatusCancelled InstanceStatus = "cancelled"
)

type WorkflowInstance struct {
	ID             uuid.UUID
	WorkflowDefID  uuid.UUID
	CurrentStateID string
	Status         InstanceStatus
	CreatedBy      string
	Context        map[string]any
	StartedAt      time.Time
	CompletedAt    *time.Time
	DeletedAt      *time.Time
}

func (i *WorkflowInstance) IsTerminal() bool {
	return i.Status == InstanceStatusCompleted || i.Status == InstanceStatusCancelled
}
