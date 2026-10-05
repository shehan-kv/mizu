package sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestSQLite(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite database: %v", err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close SQLite connection: %v", err)
		}
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping SQLite: %v", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("failed to enable SQLite foreign keys: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run SQLite migrations: %v", err)
	}

	return db
}
