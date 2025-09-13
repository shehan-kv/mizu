package postgres

import "database/sql"

// ChangeRequestStorePostgres implements the ChangeRequestStore interface using Postgres.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ChangeRequestStorePostgres struct {
	db *sql.DB
}

// NewChangeRequestStore constructs a ChangeRequestStorePostgres that persists
// ChangeRequest entities in a SQLite database via the provided *sql.DB.
func NewChangeRequestStore(db *sql.DB) *ChangeRequestStorePostgres {
	return &ChangeRequestStorePostgres{db: db}
}
