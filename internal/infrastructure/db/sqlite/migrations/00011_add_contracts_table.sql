-- +goose Up
CREATE TABLE contracts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    name TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    terms TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE
);


CREATE INDEX idx_contracts_project_id ON contracts(project_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_created_at ON contracts(created_at DESC);
CREATE INDEX idx_contracts_project_created_at ON contracts(project_id, created_at DESC);
CREATE INDEX idx_contracts_project_status ON contracts(project_id, status);
CREATE INDEX idx_contracts_project_status_stats ON contracts(project_id, status);
CREATE INDEX idx_contracts_id_version ON contracts(id, version);
CREATE INDEX idx_contracts_name ON contracts(name);


-- +goose Down
DROP INDEX IF EXISTS idx_contracts_project_id;
DROP INDEX IF EXISTS idx_contracts_status;
DROP INDEX IF EXISTS idx_contracts_created_at;
DROP INDEX IF EXISTS idx_contracts_project_created_at;
DROP INDEX IF EXISTS idx_contracts_project_status;
DROP INDEX IF EXISTS idx_contracts_project_status_stats;
DROP INDEX IF EXISTS idx_contracts_id_version;
DROP INDEX IF EXISTS idx_contracts_name;

DROP TABLE IF EXISTS contracts;