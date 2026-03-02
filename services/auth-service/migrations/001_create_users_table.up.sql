CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) NOT NULL,
    username      VARCHAR(30)  NOT NULL,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ  NULL,

    CONSTRAINT users_email_key    UNIQUE (email),
    CONSTRAINT users_username_key UNIQUE (username)
);

CREATE INDEX IF NOT EXISTS idx_users_email    ON users (email)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username) WHERE deleted_at IS NULL;