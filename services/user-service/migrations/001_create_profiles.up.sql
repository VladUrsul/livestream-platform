CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS profiles (
    user_id      UUID         PRIMARY KEY,
    username     VARCHAR(30)  NOT NULL UNIQUE,
    email        VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(60)  NOT NULL DEFAULT '',
    bio          VARCHAR(200) NOT NULL DEFAULT '',
    avatar_url   TEXT         NOT NULL DEFAULT '',
    followers    INTEGER      NOT NULL DEFAULT 0,
    following    INTEGER      NOT NULL DEFAULT 0,
    is_live      BOOLEAN      NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS follows (
    follower_id UUID        NOT NULL REFERENCES profiles(user_id) ON DELETE CASCADE,
    followee_id UUID        NOT NULL REFERENCES profiles(user_id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id)
);

CREATE INDEX IF NOT EXISTS idx_profiles_username ON profiles (username);
CREATE INDEX IF NOT EXISTS idx_profiles_is_live  ON profiles (is_live) WHERE is_live = true;
CREATE INDEX IF NOT EXISTS idx_follows_followee  ON follows (followee_id);