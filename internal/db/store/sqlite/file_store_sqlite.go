package sqlite

import "database/sql"

// FileStoreSqlite implements the FileStore interface using SQLite.
// It persists File entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type FileStoreSqlite struct {
	db *sql.DB
}

// NewFileStore constructs a FileStoreSqlite that persists
// File entities in a SQLite database via the provided *sql.DB.
func NewFileStore(db *sql.DB) *FileStoreSqlite {
	return &FileStoreSqlite{db: db}
}
