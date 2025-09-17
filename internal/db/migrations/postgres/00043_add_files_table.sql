-- +goose Up
CREATE TABLE files (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    orig_name TEXT NOT NULL,
    saved_name TEXT NOT NULL,
    url TEXT NOT NULL,
    size BIGINT NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_files_channel_id ON files(channel_id);
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_orig_name ON files(orig_name);
CREATE INDEX IF NOT EXISTS idx_files_saved_name ON files(saved_name);

-- +goose Down
DROP INDEX IF EXISTS idx_files_channel_id;
DROP INDEX IF EXISTS idx_files_user_id;
DROP INDEX IF EXISTS idx_files_orig_name;
DROP INDEX IF EXISTS idx_files_saved_name;
DROP TABLE IF EXISTS files;

