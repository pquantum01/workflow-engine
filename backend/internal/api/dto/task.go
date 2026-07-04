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
	ID                     string     `json:"id"`
	WorkflowInstanceID     string     `json:"workflow_instance_id"`
	StateID                string     `json:"state_id"`
	TaskKey                string     `json:"task_key"`
	TaskLabel              string     `json:"task_label"`
	TaskType               string     `json:"task_type"`
	Status                 string     `json:"status"`
	Outcome                *string    `json:"outcome,omitempty"`
	CreatedBy              string     `json:"created_by"`
	CreatedAt              time.Time  `json:"created_at"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	Priority               string     `json:"priority"`
	ExpectedCompletionDate *string    `json:"expected_completion_date,omitempty"`
	ExpectedCompletionTime *string    `json:"expected_completion_time,omitempty"`
	IsAdhoc                bool       `json:"is_adhoc"`
	CommentCount           int        `json:"comment_count"`
	HasAttachment          bool       `json:"has_attachment"`
}

type PatchTaskRequest struct {
	Priority               *string `json:"priority"`
	ExpectedCompletionDate *string `json:"expected_completion_date"`
	ExpectedCompletionTime *string `json:"expected_completion_time"`
}

type CreateAdhocTaskRequest struct {
	StateID                string  `json:"state_id"`
	Label                  string  `json:"label"`
	TaskType               string  `json:"task_type"`
	CreatedBy              string  `json:"created_by"`
	Priority               string  `json:"priority"`
	ExpectedCompletionDate *string `json:"expected_completion_date"`
	ExpectedCompletionTime *string `json:"expected_completion_time"`
}
