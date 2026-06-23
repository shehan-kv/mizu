-- +goose Up
CREATE TABLE channel_members (
    channel_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (channel_id, user_id),
    CONSTRAINT fk_channel_members_channel
        FOREIGN KEY (channel_id)
        REFERENCES channels(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_channel_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_channel_members_user_id ON channel_members (user_id);


-- +goose Down
DROP INDEX IF EXISTS idx_channel_members_user_id;

DROP TABLE IF EXISTS channel_members;