-- +goose Up
CREATE TABLE contract_version_statuses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_contract_version_statuses_name ON contract_version_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_version_statuses_name;
DROP TABLE IF EXISTS contract_version_statuses;
