-- +goose Up
CREATE TABLE invoice_items (
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    qty DECIMAL(19,4) NOT NULL,
    unit_price DECIMAL(19,4) NOT NULL,
    discount_rate DECIMAL(19,4) NOT NULL,
    discount_type TEXT NOT NULL,
    tax_rate DECIMAL(19,4) NOT NULL,
    tax_type TEXT NOT NULL,
    discount_amount_per_unit DECIMAL(19,4) NOT NULL,
    taxable_base_per_unit DECIMAL(19,4) NOT NULL,
    tax_amount_per_unit DECIMAL(19,4) NOT NULL,
    line_gross DECIMAL(19,4) NOT NULL,
    line_discount DECIMAL(19,4) NOT NULL,
    line_net DECIMAL(19,4) NOT NULL,
    line_tax DECIMAL(19,4) NOT NULL,
    line_total DECIMAL(19,4) NOT NULL,
    FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);


-- +goose Down
DROP INDEX IF EXISTS idx_invoice_items_invoice_id;
DROP TABLE IF EXISTS invoice_items;