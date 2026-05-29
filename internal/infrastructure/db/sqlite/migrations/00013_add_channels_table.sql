-- +goose Up
CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    project_id TEXT NULL,
    name TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_channels_project_id ON channels(project_id);


-- +goose Down
DROP INDEX IF EXISTS idx_channels_project_id;

DROP TABLE IF EXISTS channels;