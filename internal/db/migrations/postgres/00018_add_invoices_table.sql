-- +goose Up
CREATE TABLE invoices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL,
    is_invoice BOOLEAN,
    status BIGINT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, 
    due_at TIMESTAMPTZ DEFAULT NULL, 
    total DECIMAL(19,4) NOT NULL,
    discount DECIMAL(19,4) NOT NULL,
    tax DECIMAL(19,4) NOT NULL,
    currency_code TEXT, 
    note TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES invoice_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_invoices_project_id ON invoices(project_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);


-- +goose Down
DROP INDEX IF EXISTS idx_invoices_project_id;
DROP INDEX IF EXISTS idx_invoices_status;
DROP TABLE IF EXISTS invoices;
