-- Rooms table
CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_number VARCHAR(10)  NOT NULL UNIQUE,
    floor       INT          NOT NULL,
    capacity    INT          NOT NULL DEFAULT 2,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Residents table (links user_id to a room)
CREATE TABLE IF NOT EXISTS residents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL UNIQUE,
    room_id      UUID        NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    role         VARCHAR(20) NOT NULL DEFAULT 'student' CHECK (role IN ('student', 'manager', 'admin')),
    check_in_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_out_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_residents_room_id ON residents(room_id);
CREATE INDEX IF NOT EXISTS idx_residents_user_id ON residents(user_id);
