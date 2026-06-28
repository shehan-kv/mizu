-- +goose Up
CREATE TABLE recoveries (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recoveries_token ON recoveries(token);
CREATE INDEX idx_recoveries_created_at ON recoveries(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_recoveries_created_at;
DROP INDEX IF EXISTS idx_recoveries_token;
DROP TABLE IF EXISTS recoveries;