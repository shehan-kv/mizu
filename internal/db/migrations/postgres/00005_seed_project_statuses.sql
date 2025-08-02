-- +goose Up
INSERT INTO project_statuses (name) VALUES 
('started'),
('paused'),
('cancelled'),
('completed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM project_statuses WHERE name IN ('started', 'paused', 'cancelled', 'completed');
