-- +goose Up
CREATE TABLE contract_versions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    contract_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status BIGINT NOT NULL,
    version TEXT NOT NULL,
    contract TEXT NOT NULL,
    UNIQUE (contract_id, version),
    FOREIGN KEY (status) REFERENCES contract_version_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contract_versions_contract_id ON contract_versions(contract_id);
CREATE INDEX IF NOT EXISTS idx_contract_versions_version ON contract_versions(version);

-- +goose Down
DROP INDEX IF EXISTS idx_contract_versions_contract_id;
DROP INDEX IF EXISTS idx_contract_versions_version;
DROP TABLE IF EXISTS contract_versions;