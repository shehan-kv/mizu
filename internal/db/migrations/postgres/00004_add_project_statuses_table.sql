-- +goose Up
CREATE TABLE project_statuses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_project_statuses_name ON project_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_project_statuses_name;
DROP TABLE IF EXISTS project_statuses;
