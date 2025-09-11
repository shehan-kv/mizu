-- +goose Up
CREATE TABLE change_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    req_user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status INTEGER NOT NULL,
    title TEXT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (req_user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES change_request_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_change_requests_project_id ON change_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_change_requests_req_user_id ON change_requests(req_user_id);
CREATE INDEX IF NOT EXISTS idx_change_requests_status ON change_requests(status);
CREATE INDEX IF NOT EXISTS idx_change_requests_title ON change_requests(title);

-- +goose Down
DROP INDEX IF EXISTS idx_change_requests_project_id;
DROP INDEX IF EXISTS idx_change_requests_req_user_id;
DROP INDEX IF EXISTS idx_change_requests_status;
DROP INDEX IF EXISTS idx_change_requests_title;
DROP TABLE IF EXISTS change_requests;