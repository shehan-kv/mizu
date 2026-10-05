-- +goose Up
CREATE TABLE invoice_items (
    invoice_id UUID NOT NULL,
    description TEXT NOT NULL,
    qty NUMERIC(19,4) NOT NULL,
    unit_price NUMERIC(19,4) NOT NULL,
    discount_rate NUMERIC(19,4) NOT NULL,
    discount_type TEXT NOT NULL,
    tax_rate NUMERIC(19,4) NOT NULL,
    tax_type TEXT NOT NULL,
    discount_amount_per_unit NUMERIC(19,4) NOT NULL,
    taxable_base_per_unit NUMERIC(19,4) NOT NULL,
    tax_amount_per_unit NUMERIC(19,4) NOT NULL,
    line_gross NUMERIC(19,4) NOT NULL,
    line_discount NUMERIC(19,4) NOT NULL,
    line_net NUMERIC(19,4) NOT NULL,
    line_tax NUMERIC(19,4) NOT NULL,
    line_total NUMERIC(19,4) NOT NULL,
    CONSTRAINT fk_invoice_items_invoice
        FOREIGN KEY (invoice_id)
        REFERENCES invoices(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE INDEX idx_invoice_items_invoice_id ON invoice_items (invoice_id);

-- +goose Down
DROP INDEX IF EXISTS idx_invoice_items_invoice_id;

DROP TABLE IF EXISTS invoice_items;