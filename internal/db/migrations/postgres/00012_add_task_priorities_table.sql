-- +goose Up
CREATE TABLE task_priorities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_task_priorities_name ON task_priorities(name);


-- +goose Down
DROP INDEX IF EXISTS idx_task_priorities_name;
DROP TABLE IF EXISTS task_priorities;
