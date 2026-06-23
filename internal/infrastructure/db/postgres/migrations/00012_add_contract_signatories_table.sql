-- +goose Up
CREATE TABLE contract_signatories (
    contract_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (contract_id, user_id),
    CONSTRAINT fk_contract_signatories_contract
        FOREIGN KEY (contract_id)
        REFERENCES contracts(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_contract_signatories_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE INDEX idx_contract_signatories_user_id ON contract_signatories (user_id);
CREATE INDEX idx_contract_signatories_user_contract ON contract_signatories (user_id, contract_id);
CREATE INDEX idx_contract_signatories_status ON contract_signatories (status);


-- +goose Down
DROP INDEX IF EXISTS idx_contract_signatories_user_id;
DROP INDEX IF EXISTS idx_contract_signatories_user_contract;
DROP INDEX IF EXISTS idx_contract_signatories_status;

DROP TABLE IF EXISTS contract_signatories;