-- +goose Up
INSERT INTO task_statuses (name) VALUES 
('backlog'),
('in-progress'),
('completed')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM task_statuses WHERE name IN ('backlog', 'in-progress', 'completed',);
