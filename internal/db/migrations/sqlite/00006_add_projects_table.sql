-- +goose Up
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    status INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status) REFERENCES project_statuses(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);


-- +goose Down
DROP INDEX IF EXISTS idx_projects_name;
DROP TABLE IF EXISTS projects;
