package sqlite

import "database/sql"

// ChangeRequestStoreSqlite implements the ChangeRequestStore interface using SQLite.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ChangeRequestStoreSqlite struct {
	db *sql.DB
}

// NewChangeRequestStore constructs a ChangeRequestStoreSqlite that persists
// ChangeRequest entities in a SQLite database via the provided *sql.DB.
func NewChangeRequestStore(db *sql.DB) *ChangeRequestStoreSqlite {
	return &ChangeRequestStoreSqlite{db: db}
}
