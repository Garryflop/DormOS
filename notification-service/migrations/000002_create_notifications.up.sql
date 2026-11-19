CREATE TABLE IF NOT EXISTS notifications (
                                             id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL,
    title      VARCHAR(255) NOT NULL,
    message    TEXT         NOT NULL,
    is_read    BOOLEAN      DEFAULT FALSE,
    created_at TIMESTAMPTZ  DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_notifications_user_id  ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read  ON notifications(user_id, is_read);
