package postgres

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/lib/pq"
)

// ChangeRequestStorePostgres implements the ChangeRequestStore interface using Postgres.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ChangeRequestStorePostgres struct {
	db *sql.DB
}

// NewChangeRequestStore constructs a ChangeRequestStorePostgres that persists
// ChangeRequest entities in a SQLite database via the provided *sql.DB.
func NewChangeRequestStore(db *sql.DB) *ChangeRequestStorePostgres {
	return &ChangeRequestStorePostgres{db: db}
}

func (q *ChangeRequestStorePostgres) CreateOne(
	ctx context.Context,
	userId int64,
	arg *params.ChangeRequestCreate) (int64, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, store.ErrInsertFailed
	}

	defer tx.Rollback()

	reqQuery := `
	INSERT INTO change_requests(project_id, req_user_id, title, status)
	VALUES($1, $2, $3, (SELECT id FROM change_request_statuses WHERE name = $4)) RETURNING id
	`

	var reqId int64
	err = tx.QueryRowContext(
		ctx,
		reqQuery,
		arg.ProjectId,
		userId,
		arg.Title,
		params.ChangeRequestInProgress,
	).Scan(&reqId)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
				return 0, store.ErrNotNullViolation
			}

			return 0, store.ErrInsertFailed
		}
	}

	reqEntryQuery := `
	INSERT INTO change_request_entries(request_id, user_id, content)
	VALUES($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, reqEntryQuery, reqId, userId, arg.Content)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == "foreign_key_violation" {
				return 0, store.ErrForeignKeyViolation
			}

			if err.Code.Name() == "not_null_violation" {
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
