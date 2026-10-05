-- +goose Up
CREATE TABLE project_members (
    project_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (project_id, user_id),
    CONSTRAINT fk_project_members_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_project_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_project_members_user_id ON project_members(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_project_members_user_id;

DROP TABLE IF EXISTS project_members;