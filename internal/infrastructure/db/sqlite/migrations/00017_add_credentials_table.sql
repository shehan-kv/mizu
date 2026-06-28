-- +goose Up
CREATE TABLE credentials (
    user_id TEXT PRIMARY KEY,
    hash TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS credentials;