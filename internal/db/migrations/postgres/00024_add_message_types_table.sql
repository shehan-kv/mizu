-- +goose Up
CREATE TABLE message_types (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_message_types_name ON message_types(name);


-- +goose Down
DROP INDEX IF EXISTS idx_message_types_name;
DROP TABLE IF EXISTS message_types;
