-- +goose Up
CREATE TABLE contract_signatures (
    version_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status BIGINT NOT NULL,
    FOREIGN KEY (version_id) REFERENCES contract_versions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status) REFERENCES contract_signature_statuses(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (version_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_contract_signatures_version_id ON contract_signatures(version_id);
CREATE INDEX IF NOT EXISTS idx_contract_signatures_user_id ON contract_signatures(user_id);
CREATE INDEX IF NOT EXISTS idx_contract_signatures_status ON contract_signatures(status);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_signatures_version_id;
DROP INDEX IF EXISTS idx_contract_signatures_user_id;
DROP INDEX IF EXISTS idx_contract_signatures_status;
DROP TABLE IF EXISTS contract_signatures;
