CREATE TABLE IF NOT EXISTS telegram_notifications (
    id SERIAL PRIMARY KEY,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'queued',
    target_user_id INT NULL REFERENCES users(id) ON DELETE SET NULL,
    entity_id INT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    enqueued_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    sent_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT telegram_notifications_type_check CHECK (type IN ('weekly_pack_results', 'pending_user', 'movie_proposal')),
    CONSTRAINT telegram_notifications_status_check CHECK (status IN ('queued', 'sent', 'failed')),
    CONSTRAINT telegram_notifications_attempts_check CHECK (attempts >= 0)
);

CREATE INDEX IF NOT EXISTS idx_telegram_notifications_status_created
    ON telegram_notifications(status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_telegram_notifications_type_created
    ON telegram_notifications(type, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_telegram_notifications_target_created
    ON telegram_notifications(target_user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_telegram_notifications_enqueued_by_created
    ON telegram_notifications(enqueued_by_user_id, created_at DESC, id DESC);
