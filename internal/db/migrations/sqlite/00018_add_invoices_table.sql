-- +goose Up
CREATE TABLE invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    is_invoice BOOLEAN,
    status INTEGER NOT NULL,
    issued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_at DATETIME DEFAULT NULL, 
    total DECIMAL(19,4) NOT NULL,
    discount DECIMAL(19,4) NOT NULL,
    tax DECIMAL(19,4) NOT NULL,
    currency_code TEXT, 
    note TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES invoice_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);


-- +goose Down
DROP TABLE IF EXISTS invoices;
