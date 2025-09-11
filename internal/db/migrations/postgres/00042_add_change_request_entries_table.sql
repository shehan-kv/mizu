-- +goose Up
CREATE TABLE change_request_entries (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    request_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    content TEXT NOT NULL,
    FOREIGN KEY (request_id) REFERENCES change_requests(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
);

CREATE INDEX IF NOT EXISTS idx_change_request_entries_request_id ON change_request_entries(request_id);
CREATE INDEX IF NOT EXISTS idx_change_request_entries_user_id ON change_request_entries(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_change_request_entries_request_id;
DROP INDEX IF EXISTS idx_change_request_entries_user_id;
DROP TABLE IF EXISTS change_request_entries;