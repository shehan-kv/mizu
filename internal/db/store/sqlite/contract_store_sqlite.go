package sqlite

import "database/sql"

// ContractStoreSqlite implements the ContractStore interface using SQLite.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ContractStoreSqlite struct {
	db *sql.DB
}

// NewContractStore constructs a ContractStoreSqlite that persists
// Contract entities in a SQLite database via the provided *sql.DB.
func NewContractStore(db *sql.DB) *ContractStoreSqlite {
	return &ContractStoreSqlite{db: db}
}
