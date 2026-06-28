-- +goose Up
CREATE TABLE recoveries (
    user_id TEXT PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

CREATE INDEX idx_recoveries_token ON recoveries(token);
CREATE INDEX idx_recoveries_created_at ON recoveries(created_at);
CREATE INDEX idx_recoveries_expires_at ON recoveries(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_recoveries_token;
DROP INDEX IF EXISTS idx_recoveries_created_at;
DROP INDEX IF EXISTS idx_recoveries_expires_at;
DROP TABLE IF EXISTS recoveries;