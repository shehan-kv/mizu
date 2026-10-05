-- +goose Up
CREATE TABLE task_assignees (
    task_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (task_id, user_id),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_task_assignees_task_id ON task_assignees(task_id);
CREATE INDEX idx_task_assignees_user_id ON task_assignees(user_id);



-- +goose Down
DROP INDEX IF EXISTS idx_task_assignees_task_id;
DROP INDEX IF EXISTS idx_task_assignees_user_id;
DROP TABLE IF EXISTS task_assignees;