-- +goose Up
CREATE TABLE contract_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    contract_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    version TEXT NOT NULL UNIQUE,
    contract TEXT NOT NULL,
    FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contract_versions_contract_id ON contract_versions(contract_id);
CREATE INDEX IF NOT EXISTS idx_contract_versions_version ON contract_versions(version);

-- +goose Down
DROP INDEX IF EXISTS idx_contract_versions_contract_id;
DROP INDEX IF EXISTS idx_contract_versions_version;
DROP TABLE IF EXISTS contract_versions;