-- +goose Up
CREATE TABLE files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    orig_name TEXT NOT NULL,
    saved_name TEXT NOT NULL,
    url TEXT NOT NULL,
    size INTEGER NOT NULL,
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

