package sqlite

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func RunMigrations(db *sql.DB) error {

	goose.SetLogger(goose.NopLogger())
	goose.SetTableName("mizu_db_version")

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return fmt.Errorf("setting migration dialect failed: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	return nil
}
