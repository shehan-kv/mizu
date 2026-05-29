-- +goose Up
CREATE TABLE currencies (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimal_places INTEGER NOT NULL DEFAULT 2
);

CREATE INDEX idx_currencies_name ON currencies(name);
CREATE INDEX idx_currencies_code ON currencies(code);

-- +goose Down
DROP INDEX IF EXISTS idx_currencies_name;
DROP INDEX IF EXISTS idx_currencies_code;
DROP TABLE IF EXISTS currencies;