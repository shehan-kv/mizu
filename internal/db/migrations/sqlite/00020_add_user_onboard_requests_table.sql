-- +goose Up
CREATE TABLE user_verify_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_valid BOOLEAN,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_verify_requests_token ON user_verify_requests(token);


-- +goose Down
DROP INDEX IF EXISTS idx_user_verify_requests_token;
DROP TABLE IF EXISTS user_verify_requests;
