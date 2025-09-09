-- +goose Up
INSERT INTO invoice_history_events (name) VALUES 
('created'),
('status_changed'),
('converted'),
('accepted'),
('rejected'),
('cancelled'),
('paid'),
('emailed'),
('downloaded')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM invoice_history_events WHERE name IN (
    'created', 
    'status_changed', 
    'converted',
    'accepted',
    'rejected',
    'cancelled',
    'paid',
    'emailed',
    'downloaded'
);
