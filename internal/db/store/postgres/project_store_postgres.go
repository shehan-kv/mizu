package postgres

import (
	"context"
	"database/sql"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/errdefs"

	"github.com/mattn/go-sqlite3"
)

// Postgres implementation of UserStore interface
type ProjectStorePostgres struct {
	db *sql.DB
}

// Creates a new instance of a ProjectStorePostgres
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *ProjectStorePostgres
func NewProjectStore(db *sql.DB) *ProjectStorePostgres {
	return &ProjectStorePostgres{db: db}
}

func (q *ProjectStorePostgres) CreateOne(ctx context.Context, arg *params.ProjectCreateParams) (int64, error) {

	query := `
	INSERT INTO projects(name, status) 
	VALUE(?, (SELECT id FROM project_statuses WHERE name = ?)) RETURNING id
	`

	var id int64 = 0
	err := q.db.QueryRowContext(
		ctx,
		query,
		arg.Name,
		arg.Status,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sqlite3.ErrConstraintUnique) {
			return id, errdefs.ErrDbUniqueViolation
		}
		return id, errdefs.ErrDbInsertFailed
	}

	return id, nil
}
