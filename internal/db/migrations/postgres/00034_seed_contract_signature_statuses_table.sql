-- +goose Up
INSERT INTO contract_signature_statuses (name) VALUES 
('pending'),
('rejected'),
('signed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM contract_signature_statuses WHERE name IN ('pending', 'rejected', 'signed');

