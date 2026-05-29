-- +goose Up
CREATE TABLE verifications (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_verifications_user_id ON verifications(user_id);
CREATE INDEX idx_verifications_created_at ON verifications(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_verifications_user_id;
DROP INDEX IF EXISTS idx_verifications_created_at;
DROP TABLE IF EXISTS verifications;