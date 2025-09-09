-- +goose Up
CREATE TABLE invoice_history_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_invoice_history_events_name ON invoice_history_events(name);


-- +goose Down
DROP INDEX IF EXISTS idx_invoice_history_events_name;
DROP TABLE IF EXISTS invoice_history_events;
