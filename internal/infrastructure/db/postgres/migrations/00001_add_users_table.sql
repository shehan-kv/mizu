-- +goose Up
CREATE TABLE users (
    id UUID NOT NULL DEFAULT uuidv7(),
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    title TEXT,
    email TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    image TEXT,
    image_mime TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_signin_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,

    CONSTRAINT users_pk PRIMARY KEY (id),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_role_active ON users(role, is_active);
CREATE INDEX idx_users_created_at ON users(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_role_active;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;

DROP TABLE IF EXISTS users;