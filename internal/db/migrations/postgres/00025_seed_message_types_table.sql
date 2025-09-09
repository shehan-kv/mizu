-- +goose Up
INSERT INTO message_types (name) VALUES 
('user'),
('quote'),
('invoice'),
('contract'),
('file-upload'),
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM message_types WHERE name IN ('user', 'quote', 'invoice', 'contract', 'file-upload');

