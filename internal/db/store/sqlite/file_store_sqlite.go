package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
)

// FileStoreSqlite implements the FileStore interface using SQLite.
// It persists File entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type FileStoreSqlite struct {
	db *sql.DB
}

// NewFileStore constructs a FileStoreSqlite that persists
// File entities in a SQLite database via the provided *sql.DB.
func NewFileStore(db *sql.DB) *FileStoreSqlite {
	return &FileStoreSqlite{db: db}
}

func (q *FileStoreSqlite) CreateOne(ctx context.Context, arg params.FileCreate) (int64, error) {

	query := `
	INSERT INTO files(channel_id, user_id, orig_name, saved_name, url, size)
	VALUES(?, ?, ?, ?, ?, ?) RETURNING id
	`

	var fileId int64
	err := q.db.QueryRowContext(
		ctx,
		query,
		arg.ChannelId,
		arg.UserId,
		arg.OriginalName,
		arg.SavedName,
		arg.Url,
		arg.Size,
	).Scan(&fileId)

	if err != nil {
		return 0, store.ErrInsertFailed
	}

	return fileId, nil
}
