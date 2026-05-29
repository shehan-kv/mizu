-- +goose Up
CREATE TABLE channel_members (
    channel_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_channel_members_user_id ON channel_members(user_id);


-- +goose Down
DROP INDEX IF EXISTS idx_channel_members_user_id;

DROP TABLE IF EXISTS channel_members;