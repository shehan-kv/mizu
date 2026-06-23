-- +goose Up
CREATE TABLE verifications (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verifications_user_id ON verifications(user_id);
CREATE INDEX idx_verifications_created_at ON verifications(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_verifications_created_at;
DROP INDEX IF EXISTS idx_verifications_user_id;

DROP TABLE IF EXISTS verifications;