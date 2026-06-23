-- +goose Up
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    project_id UUID,
    name TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_channels_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_channels_project_id ON channels (project_id);


-- +goose Down
DROP INDEX IF EXISTS idx_channels_project_id;

DROP TABLE IF EXISTS channels;