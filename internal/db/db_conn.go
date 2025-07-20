package db

import (
	"database/sql"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"mizu/internal/db/migrations"
	"mizu/internal/db/store"
	"mizu/internal/db/store/postgres"
	"mizu/internal/db/store/sqlite"
	"mizu/internal/logger"
)

// Holds info and methods needed to connect
// to databases
type DbConn struct {
	logger     logger.Logger
	sqlDb      *sql.DB
	dbEngine   string
	connString string
}

// Creates a new instance of a DbConn
//
// Parameters:
//   - an implementation of logger.Logger
//
// Returns:
//   - *DbConn
func New(l logger.Logger) *DbConn {
	return &DbConn{logger: l}
}

// Initializes the database connection based on the DB_DRIVER
// and DB_CONNECTION_STRING environment variables.
// Runs migrations once connected.
func (dbs *DbConn) Init() {
	dbs.dbEngine = strings.TrimSpace(strings.ToLower(os.Getenv("DB_DRIVER")))
	dbs.connString = strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_STRING")))

	switch dbs.dbEngine {
	case "sqlite":
		dbs.logger.Info("SQLite selected")

		if dbs.connString == "" {
			dbs.logger.Info("SQLite connection string not found")
			dbs.logger.Info("Using default SQLite database")
			dbs.sqlDb = dbs.connectToSQL("sqlite3", "file:mizu.db")
		}

		dbs.logger.Info("SQLite connection string found")
		dbs.sqlDb = dbs.connectToSQL("sqlite3", dbs.connString)

		// Run migrations
		migrations.MigrateSQLite(dbs.sqlDb, dbs.logger)

	case "postgres":
		dbs.logger.Info("PostgreSQL selected")
		if dbs.connString == "" {
			dbs.logger.Fatal("PostgreSQL connection string not found")
		}

		dbs.sqlDb = dbs.connectToSQL("postgres", dbs.connString)

		// Run migrations
		migrations.MigratePostgres(dbs.sqlDb, dbs.logger)

	default:
		dbs.logger.Info("No valid database driver selected")
		dbs.logger.Info("Using SQLite with default settings")

		dbs.sqlDb = dbs.connectToSQL("sqlite3", "file:mizu.db")

		// Run migrations
		migrations.MigrateSQLite(dbs.sqlDb, dbs.logger)
	}
}

// Connects to an SQL database.
//
// Parameters:
//   - engine: the database engine to connect to
//   - connString: connection string to connect to the database
//
// Returns:
//   - *sql.DB
func (store *DbConn) connectToSQL(engine string, connString string) *sql.DB {

	store.logger.Info("Connecting to database...")

	db, err := sql.Open(engine, connString)
	if err != nil {
		store.logger.Fatal("Could not open database connection, check connection string")
	}

	err = db.Ping()
	if err != nil {
		store.logger.Fatal("Could not connect to database")
	}

	store.logger.Info("Connected to database")
	return db
}

// Creates a new instance of a UserStore.
//
// Returns:
//   - a pointer to an implementation of UserStore
func (dbs *DbConn) NewUserStore() store.UserStore {
	switch dbs.dbEngine {
	case "sqlite":
		return sqlite.NewUserStore(dbs.sqlDb)

	case "postgres":
		return postgres.NewUserStore(dbs.sqlDb)

	default:
		return sqlite.NewUserStore(dbs.sqlDb)
	}
}

// Closes the database connection
func (dbs *DbConn) Close() {
	dbs.sqlDb.Close()
}
