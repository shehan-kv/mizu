-- +goose Up
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    channel_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_messages_channel
        FOREIGN KEY (channel_id)
        REFERENCES channels(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender
        FOREIGN KEY (sender_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_messages_channel_id_created_at ON messages (channel_id, created_at);
CREATE INDEX idx_messages_sender_id ON messages (sender_id);


-- +goose Down
DROP INDEX IF EXISTS idx_messages_channel_id_created_at;
DROP INDEX IF EXISTS idx_messages_sender_id;

DROP TABLE IF EXISTS messages;