-- +goose Up
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    priority INTEGER NOT NULL,
    status INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estimated_time_minutes INTEGER NOT NULL,
    UNIQUE (project_id, name),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES task_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (priority) REFERENCES task_priorities(id) ON DELETE RESTRICT ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tasks_name ON tasks(name);


-- +goose Down
DROP INDEX IF EXISTS idx_tasks_name;
DROP TABLE IF EXISTS tasks;