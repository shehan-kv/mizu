-- +goose Up
CREATE TABLE invoice_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    qty DECIMAL(19, 4) NOT NULL,
    unit_price DECIMAL(19, 4) NOT NULL,
    unit_discount DECIMAL(19, 4) NOT NULL,
    discount_type TEXT NOT NULL CHECK (discount_type IN ("fixed", "percentage")),
    unit_tax DECIMAL(19, 4) NOT NULL,
    tax_type TEXT NOT NULL CHECK (tax_type IN ("fixed", "percentage")),
    tax DECIMAL(19, 4) NOT NULL,
    discount DECIMAL(19, 4) NOT NULL,
    total DECIMAL(19, 4) NOT NULL,
    FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE ON UPDATE CASCADE
);


-- +goose Down
DROP TABLE IF EXISTS invoice_items;
