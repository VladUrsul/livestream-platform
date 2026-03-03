CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS streams (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID         NOT NULL,
    username     VARCHAR(30)  NOT NULL,
    title        VARCHAR(120) NOT NULL DEFAULT '',
    category     VARCHAR(60)  NOT NULL DEFAULT 'General',
    stream_key   TEXT         NOT NULL DEFAULT '',
    status       VARCHAR(20)  NOT NULL DEFAULT 'offline',
    viewer_count INTEGER      NOT NULL DEFAULT 0,
    started_at   TIMESTAMPTZ  NULL,
    ended_at     TIMESTAMPTZ  NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stream_keys (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL,
    key        TEXT        NOT NULL UNIQUE,
    active     BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_streams_user_id  ON streams (user_id);
CREATE INDEX IF NOT EXISTS idx_streams_username ON streams (username);
CREATE INDEX IF NOT EXISTS idx_streams_status   ON streams (status);
CREATE INDEX IF NOT EXISTS idx_stream_keys_user ON stream_keys (user_id) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_stream_keys_key  ON stream_keys (key)     WHERE active = true;