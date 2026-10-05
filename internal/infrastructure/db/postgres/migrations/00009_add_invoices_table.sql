-- +goose Up
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    project_id UUID NOT NULL,
    is_invoice BOOLEAN NOT NULL,
    status TEXT NOT NULL,
    due_at TIMESTAMPTZ,
    currency_code VARCHAR(3) NOT NULL,
    note TEXT,
    total_tax NUMERIC(19,4) NOT NULL,
    total_discount NUMERIC(19,4) NOT NULL,
    sub_total NUMERIC(19,4) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    CONSTRAINT fk_invoices_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_invoices_currency
        FOREIGN KEY (currency_code)
        REFERENCES currencies(code)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);

CREATE INDEX idx_invoices_project_id ON invoices (project_id);
CREATE INDEX idx_invoices_status ON invoices (status);
CREATE INDEX idx_invoices_is_invoice ON invoices (is_invoice);
CREATE INDEX idx_invoices_created_at ON invoices (created_at DESC);
CREATE INDEX idx_invoices_project_created_at ON invoices (project_id, created_at DESC);
CREATE INDEX idx_invoices_project_status ON invoices (project_id, status);
CREATE INDEX idx_invoices_project_is_invoice ON invoices (project_id, is_invoice);
CREATE INDEX idx_invoices_project_status_updated_at ON invoices (project_id, status, updated_at);
CREATE INDEX idx_invoices_invoice_status_currency ON invoices (is_invoice, status, currency_code);


-- +goose Down
DROP INDEX IF EXISTS idx_invoices_project_id;
DROP INDEX IF EXISTS idx_invoices_status;
DROP INDEX IF EXISTS idx_invoices_is_invoice;
DROP INDEX IF EXISTS idx_invoices_created_at;
DROP INDEX IF EXISTS idx_invoices_project_created_at;
DROP INDEX IF EXISTS idx_invoices_project_status;
DROP INDEX IF EXISTS idx_invoices_project_is_invoice;
DROP INDEX IF EXISTS idx_invoices_project_status_updated_at;
DROP INDEX IF EXISTS idx_invoices_invoice_status_currency;

DROP TABLE IF EXISTS invoices;