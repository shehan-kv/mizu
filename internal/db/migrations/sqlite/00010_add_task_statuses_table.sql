-- +goose Up
CREATE TABLE task_statuses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_task_statuses_name ON task_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_task_statuses_name;
DROP TABLE IF EXISTS task_statuses;
