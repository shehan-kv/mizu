-- +goose Up
INSERT INTO roles (name) VALUES 
('administrator'),
('client')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM roles WHERE name IN ('administrator', 'client');
