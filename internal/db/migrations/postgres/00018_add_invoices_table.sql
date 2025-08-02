-- +goose Up
CREATE TABLE invoices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL,
    is_invoice BOOLEAN,
    status BIGINT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- fix it 
    due_at TIMESTAMPTZ DEFAULT NULL, 
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
