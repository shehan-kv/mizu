-- +goose Up
CREATE TABLE project_members (
    project_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (project_id, user_id),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_project_id ON project_members(project_id);

-- +goose Down
DROP INDEX IF EXISTS idx_project_members_user_id;
DROP INDEX IF EXISTS idx_project_members_project_id;
DROP TABLE IF EXISTS project_members;
