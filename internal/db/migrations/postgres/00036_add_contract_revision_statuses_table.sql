-- +goose Up
CREATE TABLE contract_revision_statuses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_contract_revision_statuses_name ON contract_revision_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_revision_statuses_name;
DROP TABLE IF EXISTS contract_revision_statuses;
