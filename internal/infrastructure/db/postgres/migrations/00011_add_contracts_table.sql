-- +goose Up
CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    project_id UUID NOT NULL,
    name TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    terms TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    CONSTRAINT fk_contracts_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE INDEX idx_contracts_project_id ON contracts (project_id);
CREATE INDEX idx_contracts_status ON contracts (status);
CREATE INDEX idx_contracts_created_at ON contracts (created_at DESC);
CREATE INDEX idx_contracts_project_created_at ON contracts (project_id, created_at DESC);
CREATE INDEX idx_contracts_project_status ON contracts (project_id, status);


-- +goose Down
DROP INDEX IF EXISTS idx_contracts_project_id;
DROP INDEX IF EXISTS idx_contracts_status;
DROP INDEX IF EXISTS idx_contracts_created_at;
DROP INDEX IF EXISTS idx_contracts_project_created_at;
DROP INDEX IF EXISTS idx_contracts_project_status;

DROP TABLE IF EXISTS contracts;