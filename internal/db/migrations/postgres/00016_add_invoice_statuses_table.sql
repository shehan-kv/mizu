-- +goose Up
CREATE TABLE invoice_statuses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_invoice_statuses_name ON invoice_statuses(name);


-- +goose Down
DROP INDEX IF EXISTS idx_invoice_statuses_name;
DROP TABLE IF EXISTS invoice_statuses;
