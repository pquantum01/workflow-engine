CREATE TABLE webhook_outbox (
    id              UUID PRIMARY KEY,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    target_url      TEXT NOT NULL,
    attempts        INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX webhook_outbox_pending_idx
    ON webhook_outbox (next_attempt_at)
    WHERE delivered_at IS NULL AND failed_at IS NULL;
