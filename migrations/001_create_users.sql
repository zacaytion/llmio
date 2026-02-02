-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    email           CITEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    deactivated_at  TIMESTAMPTZ,
    key             TEXT NOT NULL UNIQUE,  -- Public URL key
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_username_format CHECK (username ~* '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(username) >= 2),
    CONSTRAINT users_name_not_empty CHECK (LENGTH(TRIM(name)) > 0)
);

-- Indexes
CREATE INDEX users_email_idx ON users(email);
CREATE INDEX users_username_idx ON users(username);
CREATE INDEX users_key_idx ON users(key);
CREATE INDEX users_created_at_idx ON users(created_at);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;

-- +goose StatementEnd
