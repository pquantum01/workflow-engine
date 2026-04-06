package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var validTaskTypes = map[string]bool{
	"manual":     true,
	"automated":  true,
	"approval":   true,
}

type Task struct {
	ID          uuid.UUID `json:"id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TaskType    string    `json:"task_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(t.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if t.WorkflowID == uuid.Nil {
		return fmt.Errorf("workflow_id is required")
	}
	if !validTaskTypes[t.TaskType] {
		return fmt.Errorf("task_type must be one of: manual, automated, approval")
	}
	return nil
}
