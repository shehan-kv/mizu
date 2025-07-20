package migrations

import (
	"database/sql"
	"embed"
	"mizu/internal/logger"

	"github.com/pressly/goose/v3"
)

//go:embed sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed postgres/*.sql
var postgresMigrations embed.FS

// Migrates SQLite database
//
// Parameters:
//   - db: *sql.DB
//   - logger: an implementation of logger.Logger
func MigrateSQLite(db *sql.DB, lg logger.Logger) {

	goose.SetLogger(goose.NopLogger())
	goose.SetTableName("mizu_db_version")

	lg.Info("Migrating SQLite...")

	goose.SetBaseFS(sqliteMigrations)
	if err := goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		lg.Fatal("Setting migration dialect failed")
	}
	if err := goose.Up(db, "sqlite"); err != nil {
		lg.Fatal("Database migration failed")
	}

	lg.Info("Migration successful")

}

// Migrates Postgres database
//
// Parameters:
//   - db: *sql.DB
//   - logger: an implementation of logger.Logger
func MigratePostgres(db *sql.DB, lg logger.Logger) {

	goose.SetLogger(goose.NopLogger())
	goose.SetTableName("mizu_db_version")

	lg.Info("Migrating SQLite...")

	goose.SetBaseFS(postgresMigrations)
	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		lg.Fatal("Setting migration dialect failed")
	}

	if err := goose.Up(db, "postgres"); err != nil {
		lg.Fatal("Database migration failed")
	}

	lg.Info("Migration successful")
}
