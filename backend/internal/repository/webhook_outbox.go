package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/workflow-engine/backend/internal/model"
)

type webhookOutboxRepo struct {
	db DB
}

func NewWebhookOutboxRepo(db DB) WebhookOutboxRepo {
	return &webhookOutboxRepo{db: db}
}

func (r *webhookOutboxRepo) WithTx(tx DB) WebhookOutboxRepo {
	return &webhookOutboxRepo{db: tx}
}

func (r *webhookOutboxRepo) BulkInsert(ctx context.Context, rows []*model.WebhookOutboxRow) error {
	for _, row := range rows {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO webhook_outbox (id, event_type, payload, target_url, next_attempt_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			row.ID, row.EventType, row.Payload, row.TargetURL, row.NextAttemptAt, row.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *webhookOutboxRepo) ListPending(ctx context.Context, now time.Time, limit int) ([]*model.WebhookOutboxRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_type, payload, target_url, attempts, next_attempt_at,
		       delivered_at, failed_at, last_error, created_at
		FROM webhook_outbox
		WHERE delivered_at IS NULL AND failed_at IS NULL AND next_attempt_at <= $1
		ORDER BY next_attempt_at
		LIMIT $2
		FOR UPDATE SKIP LOCKED`,
		now, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.WebhookOutboxRow
	for rows.Next() {
		var row model.WebhookOutboxRow
		var deliveredAt, failedAt sql.NullTime
		var lastError sql.NullString
		if err := rows.Scan(
			&row.ID, &row.EventType, &row.Payload, &row.TargetURL,
			&row.Attempts, &row.NextAttemptAt, &deliveredAt, &failedAt, &lastError, &row.CreatedAt,
		); err != nil {
			return nil, err
		}
		if deliveredAt.Valid {
			row.DeliveredAt = &deliveredAt.Time
		}
		if failedAt.Valid {
			row.FailedAt = &failedAt.Time
		}
		if lastError.Valid {
			row.LastError = &lastError.String
		}
		result = append(result, &row)
	}
	return result, rows.Err()
}

func (r *webhookOutboxRepo) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE webhook_outbox SET delivered_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}

func (r *webhookOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, lastError string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE webhook_outbox SET failed_at = $1, last_error = $2 WHERE id = $3`,
		time.Now(), lastError, id,
	)
	return err
}

func (r *webhookOutboxRepo) IncrementAttempt(ctx context.Context, id uuid.UUID, nextAttemptAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE webhook_outbox SET attempts = attempts + 1, next_attempt_at = $1 WHERE id = $2`,
		nextAttemptAt, id,
	)
	return err
}
