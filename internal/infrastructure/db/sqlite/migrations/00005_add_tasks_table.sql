-- +goose Up
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    priority TEXT NOT NULL,
    status TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    estimated_minutes INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, name),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_tasks_project_status_priority_created ON tasks(project_id, status, priority, created_at DESC);
CREATE INDEX idx_tasks_name ON tasks(name);
CREATE INDEX idx_tasks_description ON tasks(description);


-- +goose Down
DROP INDEX IF EXISTS idx_tasks_project_id;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_project_status_priority_created;
DROP INDEX IF EXISTS idx_tasks_name;
DROP INDEX IF EXISTS idx_tasks_description;
DROP TABLE IF EXISTS tasks;