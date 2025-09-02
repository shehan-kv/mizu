-- +goose Up
CREATE TABLE contract_signature_statuses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_contract_signature_statuses_name ON contract_signature_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_signature_statuses_name;
DROP TABLE IF EXISTS contract_signature_statuses;
