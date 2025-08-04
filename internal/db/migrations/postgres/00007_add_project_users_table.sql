-- +goose Up
CREATE TABLE project_users (
    user_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (user_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_project_users_user_id ON project_users(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_project_users_user_id;
DROP TABLE IF EXISTS project_users;
