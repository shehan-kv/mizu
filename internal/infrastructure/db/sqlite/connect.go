package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Connect(connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, fmt.Errorf("Could not open database connection: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Could not connect to database: %w", err)
	}

	return db, nil
}
