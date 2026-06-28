-- +goose Up
CREATE TABLE credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    hash TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
);

-- +goose Down
DROP TABLE IF EXISTS credentials;