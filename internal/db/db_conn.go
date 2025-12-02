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
	lg         logger.Logger
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
	return &DbConn{lg: l}
}

// Initializes the database connection based on the DB_DRIVER
// and DB_CONNECTION_STRING environment variables.
// Runs migrations once connected.
func (d *DbConn) Init() {
	d.dbEngine = strings.TrimSpace(strings.ToLower(os.Getenv("DB_DRIVER")))
	d.connString = strings.TrimSpace(strings.ToLower(os.Getenv("DB_CONNECTION_STRING")))

	switch d.dbEngine {
	case "sqlite":
		d.lg.Info("SQLite selected")

		if d.connString == "" {
			d.lg.Info("SQLite connection string not found")
			d.lg.Info("Using default SQLite database")
			d.sqlDb = d.connectToSQL("sqlite3", "file:mizu.db")
		}

		d.lg.Info("SQLite connection string found")
		d.sqlDb = d.connectToSQL("sqlite3", d.connString)

		// Run migrations
		migrations.MigrateSQLite(d.sqlDb, d.lg)

	case "postgres":
		d.lg.Info("PostgreSQL selected")
		if d.connString == "" {
			d.lg.Fatal("PostgreSQL connection string not found")
		}

		d.sqlDb = d.connectToSQL("postgres", d.connString)

		// Run migrations
		migrations.MigratePostgres(d.sqlDb, d.lg)

	default:
		d.lg.Info("No valid database driver selected")
		d.lg.Info("Using SQLite with default settings")

		d.sqlDb = d.connectToSQL("sqlite3", "file:mizu.db")

		// Run migrations
		migrations.MigrateSQLite(d.sqlDb, d.lg)
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
func (d *DbConn) connectToSQL(engine string, connString string) *sql.DB {

	d.lg.Info("Connecting to database...")

	db, err := sql.Open(engine, connString)
	if err != nil {
		d.lg.Fatal("Could not open database connection, check connection string")
	}

	err = db.Ping()
	if err != nil {
		d.lg.Fatal("Could not connect to database")
	}

	d.lg.Info("Connected to database")
	return db
}

// Creates a new instance of a UserStore.
//
// Returns:
//   - a pointer to an implementation of UserStore
func (d *DbConn) NewUserStore() store.UserStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewUserStore(d.sqlDb)

	case "postgres":
		return postgres.NewUserStore(d.sqlDb)

	default:
		return sqlite.NewUserStore(d.sqlDb)
	}
}

// Creates a new instance of a ProjectStore.
//
// Returns:
//   - a pointer to an implementation of ProjectStore
func (d *DbConn) NewProjectStore() store.ProjectStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewProjectStore(d.sqlDb)

	case "postgres":
		return postgres.NewProjectStore(d.sqlDb)

	default:
		return sqlite.NewProjectStore(d.sqlDb)
	}
}

// Creates a new instance of a Invoicetore.
//
// Returns:
//   - a pointer to an implementation of InvoiceStore
func (d *DbConn) NewInvoiceStore() store.InvoiceStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewInvoiceStore(d.sqlDb)

	case "postgres":
		return postgres.NewInvoiceStore(d.sqlDb)

	default:
		return sqlite.NewInvoiceStore(d.sqlDb)
	}
}

// Creates a new instance of a MessageStore.
//
// Returns:
//   - a pointer to an implementation of MessageStore
func (d *DbConn) NewMessagetore() store.MessageStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewMessageStore(d.sqlDb)

	case "postgres":
		return postgres.NewMessageStore(d.sqlDb)

	default:
		return sqlite.NewMessageStore(d.sqlDb)
	}
}

// Creates a new instance of a ContractStore.
//
// Returns:
//   - a pointer to an implementation of ContractStore
func (d *DbConn) NewContractStore() store.ContractStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewContractStore(d.sqlDb)

	case "postgres":
		return postgres.NewContractStore(d.sqlDb)

	default:
		return sqlite.NewContractStore(d.sqlDb)
	}
}

// Creates a new instance of a ChangeRequestStore.
//
// Returns:
//   - a pointer to an implementation of ChangeRequestStore
func (d *DbConn) NewChangeRequestStore() store.ChangeRequestStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewChangeRequestStore(d.sqlDb)

	case "postgres":
		return postgres.NewChangeRequestStore(d.sqlDb)

	default:
		return sqlite.NewChangeRequestStore(d.sqlDb)
	}
}

// Creates a new instance of a FileStore.
//
// Returns:
//   - a pointer to an implementation of FileStore
func (d *DbConn) NewFileStore() store.FileStore {
	switch d.dbEngine {
	case "sqlite":
		return sqlite.NewFileStore(d.sqlDb)

	case "postgres":
		return postgres.NewFileStore(d.sqlDb)

	default:
		return sqlite.NewFileStore(d.sqlDb)
	}
}

// Closes the database connection
func (d *DbConn) Close() {
	d.sqlDb.Close()
}
