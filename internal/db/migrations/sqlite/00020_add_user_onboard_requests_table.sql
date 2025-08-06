-- +goose Up
CREATE TABLE user_onboard_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    issued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_valid BOOLEAN,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_onboard_requests_token ON user_onboard_requests(token);


-- +goose Down
DROP INDEX IF EXISTS idx_user_onboard_requests_token;
DROP TABLE IF EXISTS user_onboard_requests;
