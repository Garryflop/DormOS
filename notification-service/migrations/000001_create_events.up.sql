CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS events (
                                      id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(255)    NOT NULL,
    description     TEXT,
    location        VARCHAR(255),
    event_date      TIMESTAMPTZ     NOT NULL,
    max_participants INT            DEFAULT 0,
    created_by      UUID            NOT NULL,
    created_at      TIMESTAMPTZ     DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS event_registrations (
                                                   id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id      UUID        NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id       UUID        NOT NULL,
    registered_at TIMESTAMPTZ DEFAULT NOW(),
    attended      BOOLEAN     DEFAULT FALSE,
    UNIQUE (event_id, user_id)
    );

CREATE INDEX IF NOT EXISTS idx_events_event_date     ON events(event_date);
CREATE INDEX IF NOT EXISTS idx_registrations_event   ON event_registrations(event_id);
CREATE INDEX IF NOT EXISTS idx_registrations_user    ON event_registrations(user_id);
