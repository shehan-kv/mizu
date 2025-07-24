-- +goose Up
CREATE TABLE project_users (
    user_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (user_id, project_id)
);


-- +goose Down
DROP TABLE IF EXISTS project_users;
