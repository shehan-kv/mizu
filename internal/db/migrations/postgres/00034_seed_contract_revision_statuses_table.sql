-- +goose Up
INSERT INTO contract_revision_statuses (name) VALUES 
('pending'),
('accepted'),
('rejected')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM contract_revision_statuses WHERE name IN ('pending', 'rejected', 'signed');

