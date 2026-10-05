-- +goose Up
CREATE TABLE task_assignees (
    task_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (task_id, user_id),
    CONSTRAINT fk_task_assignees_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_task_assignees_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_task_assignees_user_id ON task_assignees (user_id);


-- +goose Down
DROP INDEX IF EXISTS idx_task_assignees_user_id;

DROP TABLE IF EXISTS task_assignees;