package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Workflow struct {
	ID           uuid.UUID    `json:"id"`
	ProcessID    uuid.UUID    `json:"process_id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	StateMachine StateMachine `json:"state_machine"`
	SortOrder    int          `json:"sort_order"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func (w *Workflow) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(w.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if w.ProcessID == uuid.Nil {
		return fmt.Errorf("process_id is required")
	}
	return w.StateMachine.Validate()
}
