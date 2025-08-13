package postgres

import "database/sql"

// Postgres implementation of MessageStore interface
type MessageStorePostgres struct {
	db *sql.DB
}

// Creates a new instance of a MessageStorePostgres
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *MessageStorePostgres
func NewMessageStore(db *sql.DB) *MessageStorePostgres {
	return &MessageStorePostgres{db: db}
}
