CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS rooms (
    id         VARCHAR(30)  PRIMARY KEY,
    slow_mode  INTEGER      NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id    VARCHAR(30)  NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL,
    username   VARCHAR(30)  NOT NULL,
    content    TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages (room_id, created_at DESC);