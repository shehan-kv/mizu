-- +goose Up
CREATE TABLE contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    status INTEGER NOT NULL,
    UNIQUE (project_id, name),
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE
    FOREIGN KEY (status) REFERENCES contract_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contracts_project_id ON contracts(project_id);
CREATE INDEX IF NOT EXISTS idx_contracts_name ON contracts(name);

-- +goose Down
DROP INDEX IF EXISTS idx_contracts_project_id;
DROP INDEX IF EXISTS idx_contracts_name;
DROP TABLE IF EXISTS contracts;