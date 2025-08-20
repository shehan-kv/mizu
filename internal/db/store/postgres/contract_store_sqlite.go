package postgres

import "database/sql"

// ContractStorePostgres implements the ContractStore interface using Postgres.
// It persists Contract entities in a Postgres database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ContractStorePostgres struct {
	db *sql.DB
}

// NewContractStore constructs a ContractStorePostgres that persists
// Contract entities in a Postgres database via the provided *sql.DB.
func NewContractStore(db *sql.DB) *ContractStorePostgres {
	return &ContractStorePostgres{db: db}
}
