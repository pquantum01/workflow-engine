package model

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusNew        TaskStatus = "new"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusRejected   TaskStatus = "rejected"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

var ValidTaskStatuses = map[TaskStatus]bool{
	TaskStatusNew:        true,
	TaskStatusInProgress: true,
	TaskStatusDone:       true,
	TaskStatusRejected:   true,
	TaskStatusCancelled:  true,
}

type TaskInstance struct {
	ID                 uuid.UUID
	WorkflowInstanceID uuid.UUID
	StateID            string
	TaskKey            string
	TaskLabel          string
	TaskType           string
	Status             TaskStatus
	Outcome            *string
	Data               []byte
	CreatedBy          string
	CreatedAt          time.Time
	CompletedAt        *time.Time
	DeletedAt          *time.Time
}

func (t *TaskInstance) IsCompletable() bool {
	return t.Status == TaskStatusNew || t.Status == TaskStatusInProgress
}

func (t *TaskInstance) IsTerminal() bool {
	return t.Status == TaskStatusDone || t.Status == TaskStatusRejected || t.Status == TaskStatusCancelled
}
