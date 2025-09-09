-- +goose Up
CREATE TABLE invoice_history (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    event BIGINT NOT NULL, 
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_invoice BOOLEAN,
    last_status BIGINT,
    new_status BIGINT,
    FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (event) REFERENCES invoice_history_events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (last_status) REFERENCES invoice_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (new_status) REFERENCES invoice_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_invoice_history_invoice_id ON invoice_history(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_history_user_id ON invoice_history(user_id);
CREATE INDEX IF NOT EXISTS idx_invoice_history_event ON invoice_history(event);
CREATE INDEX IF NOT EXISTS idx_invoice_last_status ON invoice_history(last_status);
CREATE INDEX IF NOT EXISTS idx_invoice_new_status ON invoice_history(new_status);


-- +goose Down
DROP INDEX IF EXISTS idx_invoice_history_invoice_id;
DROP INDEX IF EXISTS idx_invoice_history_user_id;
DROP INDEX IF EXISTS idx_invoice_history_event;
DROP INDEX IF EXISTS idx_invoice_last_status;
DROP INDEX IF EXISTS idx_invoice_new_status;
DROP TABLE IF EXISTS invoice_history;
