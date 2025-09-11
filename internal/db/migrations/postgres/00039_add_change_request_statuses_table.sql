-- +goose Up
CREATE TABLE change_request_statuses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_change_request_statuses_name ON change_request_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_change_request_statuses_name;
DROP TABLE IF EXISTS change_request_statuses;
