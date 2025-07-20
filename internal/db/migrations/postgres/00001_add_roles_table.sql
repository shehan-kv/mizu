-- +goose Up
CREATE TABLE roles (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);


-- +goose Down
DROP INDEX IF EXISTS idx_roles_name;
DROP TABLE IF EXISTS roles;

