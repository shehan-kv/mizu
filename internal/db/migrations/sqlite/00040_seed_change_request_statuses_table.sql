-- +goose Up
INSERT INTO contract_revision_statuses (name) VALUES 
('in-progress'),
('waiting'),
('closed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM contract_revision_statuses WHERE name IN ('pending', 'rejected', 'signed');

