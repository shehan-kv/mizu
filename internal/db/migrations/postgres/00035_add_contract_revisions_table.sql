-- +goose Up
CREATE TABLE contract_revisions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    contract_id BIGINT NOT NULL,
    req_user_id BIGINT NOT NULL,
    res_user_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    status BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (req_user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (res_user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES contract_revision_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_contract_revisions_version_id ON contract_revisions(contract_id);
CREATE INDEX IF NOT EXISTS idx_contract_revisions_req_user_id ON contract_revisions(req_user_id);
CREATE INDEX IF NOT EXISTS idx_contract_revisions_res_user_id ON contract_revisions(res_user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_contract_revisions_version_id;
DROP INDEX IF EXISTS idx_contract_revisions_req_user_id;
DROP TABLE IF EXISTS contract_revisions;