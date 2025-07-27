-- +goose Up
CREATE TABLE task_assignees (
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (user_id, task_id)
);


-- +goose Down
DROP TABLE IF EXISTS task_assignees;
