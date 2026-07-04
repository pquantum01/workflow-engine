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
	// Phase 7 extensions
	Priority               string
	ExpectedCompletionDate *string
	ExpectedCompletionTime *string
	IsAdhoc                bool
}

func (t *TaskInstance) IsCompletable() bool {
	return t.Status == TaskStatusNew || t.Status == TaskStatusInProgress
}

func (t *TaskInstance) IsTerminal() bool {
	return t.Status == TaskStatusDone || t.Status == TaskStatusRejected || t.Status == TaskStatusCancelled
}

type TaskComment struct {
	ID             uuid.UUID
	TaskInstanceID uuid.UUID
	Author         string
	Content        string
	CreatedAt      time.Time
}

type TaskAttachment struct {
	ID             uuid.UUID
	TaskInstanceID uuid.UUID
	Filename       string
	ContentType    string
	SizeBytes      int
	S3URL          string
	UploadedAt     time.Time
	UpdatedAt      time.Time
}
