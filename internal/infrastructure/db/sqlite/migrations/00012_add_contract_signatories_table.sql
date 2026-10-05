-- +goose Up

CREATE TABLE contract_signatories (
    contract_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY (contract_id, user_id)
);

CREATE INDEX idx_contract_signatories_contract_id ON contract_signatories(contract_id);
CREATE INDEX idx_contract_signatories_user_id ON contract_signatories(user_id);
CREATE INDEX idx_contract_signatories_user_contract ON contract_signatories(user_id, contract_id);
CREATE INDEX idx_contract_signatories_status ON contract_signatories(status);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_signatories_contract_id;
DROP INDEX IF EXISTS idx_contract_signatories_user_id;
DROP INDEX IF EXISTS idx_contract_signatories_user_contract;
DROP INDEX IF EXISTS idx_contract_signatories_status;

DROP TABLE IF EXISTS contract_signatories;