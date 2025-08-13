package sqlite

import "database/sql"

// SQLite implementation of MessageStore interface
type MessageStoreSqlite struct {
	db *sql.DB
}

// Creates a new instance of a MessageStoreSqlite
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *MessageStoreSqlite
func NewMessageStore(db *sql.DB) *MessageStoreSqlite {
	return &MessageStoreSqlite{db: db}
}
