-- +goose Up
CREATE TABLE files (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    original_name TEXT NOT NULL,
    saved_name TEXT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    uploaded_at DATETIME NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_files_channel_id ON files(channel_id);
CREATE INDEX idx_files_uploaded_at ON files(uploaded_at DESC);
CREATE INDEX idx_files_channel_uploaded_at ON files(channel_id, uploaded_at DESC);
CREATE INDEX idx_files_original_name ON files(original_name);
CREATE INDEX idx_files_saved_name ON files(saved_name);

-- +goose Down
DROP INDEX IF EXISTS idx_files_channel_id;
DROP INDEX IF EXISTS idx_files_uploaded_at;
DROP INDEX IF EXISTS idx_files_channel_uploaded_at;
DROP INDEX IF EXISTS idx_files_original_name;
DROP INDEX IF EXISTS idx_files_saved_name;

DROP TABLE IF EXISTS files;