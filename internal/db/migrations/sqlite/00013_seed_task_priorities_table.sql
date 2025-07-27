-- +goose Up
INSERT INTO task_priorities (name) VALUES 
('high'),
('medium'),
('low')
ON CONFLICT(name) DO NOTHING;


-- +goose Down
DELETE FROM task_priorities WHERE name IN ('high', 'medium', 'low',);
