-- +goose Up
CREATE TABLE currencies (
    code VARCHAR(3) PRIMARY KEY,
    name TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimal_places INTEGER NOT NULL DEFAULT 2
);

CREATE INDEX idx_currencies_name ON currencies (name);

-- +goose Down
DROP INDEX IF EXISTS idx_currencies_name;

DROP TABLE IF EXISTS currencies;