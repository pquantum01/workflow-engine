package model

import (
	"time"

	"github.com/google/uuid"
)

type WebhookOutboxRow struct {
	ID            uuid.UUID
	EventType     string
	Payload       []byte
	TargetURL     string
	Attempts      int
	NextAttemptAt time.Time
	DeliveredAt   *time.Time
	FailedAt      *time.Time
	LastError     *string
	CreatedAt     time.Time
}
