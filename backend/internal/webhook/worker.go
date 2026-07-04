package webhook

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"workflow-engine/backend/internal/repository"
)

var backoffSchedule = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	12 * time.Hour,
}

type Worker struct {
	outboxRepo  repository.WebhookOutboxRepo
	subscribers []Subscriber
	client      *http.Client
	interval    time.Duration
	maxAttempts int
	log         *logrus.Logger
}

func NewWorker(
	outboxRepo repository.WebhookOutboxRepo,
	subscribers []Subscriber,
	interval time.Duration,
	maxAttempts int,
	log *logrus.Logger,
) *Worker {
	return &Worker{
		outboxRepo:  outboxRepo,
		subscribers: subscribers,
		client:      &http.Client{Timeout: 10 * time.Second},
		interval:    interval,
		maxAttempts: maxAttempts,
		log:         log,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	rows, err := w.outboxRepo.ListPending(ctx, time.Now().UTC(), 20)
	if err != nil {
		w.log.WithField("err", err).Error("webhook worker: list pending failed")
		return
	}
	for _, row := range rows {
		if err := w.deliver(ctx, row.ID, row.TargetURL, row.Payload, row.EventType, row.Attempts); err != nil {
			w.log.WithFields(logrus.Fields{
				"outbox_id": row.ID,
				"url":       row.TargetURL,
				"attempts":  row.Attempts,
				"err":       err,
			}).Warn("webhook delivery failed")
		}
	}
}

func (w *Worker) deliver(ctx context.Context, id uuid.UUID, url string, payload []byte, eventType string, attempts int) error {
	secret := ""
	for _, sub := range w.subscribers {
		if sub.URL == url {
			secret = sub.Secret
			break
		}
	}

	w.log.WithFields(logrus.Fields{
		"outbox_id":  id,
		"url":        url,
		"event_type": eventType,
		"attempts":   attempts,
	}).Debug("webhook delivering")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return w.handleFailure(ctx, id, attempts, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-Delivery", id.String())
	if secret != "" {
		req.Header.Set("X-Webhook-Signature", Sign(secret, payload))
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return w.handleFailure(ctx, id, attempts, err.Error())
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.log.WithFields(logrus.Fields{"outbox_id": id, "url": url, "status": resp.StatusCode}).Info("webhook delivered")
		return w.outboxRepo.MarkDelivered(ctx, id)
	}
	return w.handleFailure(ctx, id, attempts, fmt.Sprintf("HTTP %d", resp.StatusCode))
}

func (w *Worker) handleFailure(ctx context.Context, id uuid.UUID, attempts int, errMsg string) error {
	newAttempts := attempts + 1
	if newAttempts >= w.maxAttempts {
		w.log.WithFields(logrus.Fields{"outbox_id": id, "attempts": newAttempts, "err": errMsg}).Error("webhook exhausted max attempts, marking failed")
		return w.outboxRepo.MarkFailed(ctx, id, errMsg)
	}
	var delay time.Duration
	if attempts < len(backoffSchedule) {
		delay = backoffSchedule[attempts]
	} else {
		delay = 12 * time.Hour
	}
	next := time.Now().UTC().Add(delay)
	w.log.WithFields(logrus.Fields{"outbox_id": id, "attempts": newAttempts, "next_retry": next, "err": errMsg}).Warn("webhook delivery failed, will retry")
	return w.outboxRepo.IncrementAttempt(ctx, id, next)
}
