-- +goose Up
INSERT INTO contract_version_statuses (name) VALUES 
('pending'),
('rejected'),
('signed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM contract_version_statuses WHERE name IN ('pending', 'rejected', 'signed');

