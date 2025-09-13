-- +goose Up
INSERT INTO change_request_statuses (name) VALUES 
('in-progress'),
('waiting'),
('closed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM change_request_statuses WHERE name IN ('in-progress', 'waiting', 'closed');

