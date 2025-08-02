-- +goose Up
INSERT INTO invoice_statuses (name) VALUES 
('paid'),
('pending'),
('accepted'),
('rejected'),
('cancelled')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM invoice_statuses WHERE name IN (
    'paid', 
    'pending', 
    'accepted', 
    'rejected', 
    'cancelled');
