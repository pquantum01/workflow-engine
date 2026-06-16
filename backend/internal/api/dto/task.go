package dto

import (
	"time"
)

type CompleteTaskRequest struct {
	Status      string `json:"status"       validate:"required,oneof=done rejected"`
	NextState   string `json:"next_state"`
	CompletedBy string `json:"completed_by"`
	Outcome     string `json:"outcome"`
	Data        any    `json:"data"`
}

type TaskInstanceResponse struct {
	ID                 string     `json:"id"`
	WorkflowInstanceID string     `json:"workflow_instance_id"`
	StateID            string     `json:"state_id"`
	TaskKey            string     `json:"task_key"`
	TaskLabel          string     `json:"task_label"`
	TaskType           string     `json:"task_type"`
	Status             string     `json:"status"`
	Outcome            *string    `json:"outcome,omitempty"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}
