-- +goose Up
CREATE TABLE channel_users (
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (channel_id, user_id)
);


-- +goose Down
DROP TABLE IF EXISTS channel_users;
