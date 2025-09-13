package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/mattn/go-sqlite3"
)

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

func (q *ChangeRequestStoreSqlite) CreateOne(
	ctx context.Context,
	userId int64,
	arg params.ChangeRequestCreate) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	reqQuery := `
	INSERT INTO change_requests(project_id, req_user_id, title, status)
	VALUES(?, ?, ?, (SELECT id FROM change_request_statuses WHERE name = ?)) RETURNING id
	`

	var reqId int64
	err = tx.QueryRowContext(ctx, reqQuery, arg.ProjectId, userId, arg.Title, arg.Status).Scan(&reqId)
	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}

			return 0, store.ErrInsertFailed
		}
	}

	reqEntryQuery := `
	INSERT INTO change_request_entries(request_id, user_id, content)
	VALUES(?, ?, ?)
	`

	_, err = tx.ExecContext(ctx, reqEntryQuery, reqId, userId, arg.Content)
	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return 0, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return 0, store.ErrNotNullViolation
			}

			return 0, store.ErrInsertFailed
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, store.ErrInsertFailed
	}

	return reqId, nil
}
