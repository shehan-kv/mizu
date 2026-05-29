-- +goose Up
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    is_system INTEGER NOT NULL DEFAULT 0,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_channel_id_created_at ON messages(channel_id, created_at);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_channel_id_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;

DROP TABLE IF EXISTS messages;