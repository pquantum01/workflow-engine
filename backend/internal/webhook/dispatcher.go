package webhook

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/engine"
	"github.com/workflow-engine/backend/internal/model"
	"github.com/workflow-engine/backend/internal/repository"
)

type EventPayload struct {
	ID                 string         `json:"id"`
	EventType          string         `json:"event_type"`
	OccurredAt         time.Time      `json:"occurred_at"`
	WorkflowInstanceID string         `json:"workflow_instance_id"`
	Data               map[string]any `json:"data"`
}

// Dispatch writes webhook_outbox rows for each matching subscriber.
// Must be called inside the same transaction as the state change.
func Dispatch(
	ctx context.Context,
	outboxRepo repository.WebhookOutboxRepo,
	subscribers []Subscriber,
	events []engine.Event,
	defName string,
) error {
	if len(subscribers) == 0 || len(events) == 0 {
		return nil
	}

	now := time.Now().UTC()
	var rows []*model.WebhookOutboxRow

	for _, event := range events {
		payload := EventPayload{
			ID:                 uuid.New().String(),
			EventType:          event.Type,
			OccurredAt:         now,
			WorkflowInstanceID: event.WorkflowInstanceID.String(),
			Data:               event.Data,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		for _, sub := range subscribers {
			if !sub.Matches(event.Type, defName) {
				continue
			}
			rows = append(rows, &model.WebhookOutboxRow{
				ID:            uuid.New(),
				EventType:     event.Type,
				Payload:       payloadJSON,
				TargetURL:     sub.URL,
				NextAttemptAt: now,
				CreatedAt:     now,
			})
		}
	}

	if len(rows) == 0 {
		return nil
	}
	return outboxRepo.BulkInsert(ctx, rows)
}
