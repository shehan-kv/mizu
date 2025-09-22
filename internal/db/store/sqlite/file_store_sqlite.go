package sqlite

import (
	"context"
	"database/sql"
	"log"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strings"
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

func (q *FileStoreSqlite) GetByChannelId(
	ctx context.Context,
	channelId int64,
	arg *params.FileSearch) (*agg.WithCount[agg.FileWithUser], error) {

	var fileQuery strings.Builder
	fileQuery.WriteString(`
	SELECT 
		f.id,
		f.channel_id,
		u.id AS user_id,
		u.first_name,
		u.last_name,
		f.orig_name,
		f.saved_name,
		f.uploaded_at,
		f.url,
		f.size
	FROM files f 
	JOIN users u ON u.id = f.user_id
	WHERE f.channel_id = ?
	`)

	var countQuery strings.Builder
	countQuery.WriteString("SELECT COUNT(id) FROM files WHERE channel_id = ?")

	queryArgs := []any{channelId}
	countArgs := []any{channelId}

	if len(arg.Keyword) != 0 {
		fileQuery.WriteString(" AND f.orig_name LIKE ?")
		countQuery.WriteString(" AND orig_name LIKE ?")

		keyword := "%" + arg.Keyword + "%"

		queryArgs = append(queryArgs, keyword)
		countArgs = append(countArgs, keyword)
	}

	fileQuery.WriteString(" ORDER BY f.uploaded_at DESC")
	fileQuery.WriteString(" LIMIT ? OFFSET ?")
	queryArgs = append(queryArgs, arg.Limit, arg.Offset)

	var totalFiles int64
	err := q.db.QueryRowContext(ctx, countQuery.String(), countArgs...).Scan(&totalFiles)
	if err != nil {
		log.Println(err)
		return nil, store.ErrQueryFailed
	}

	rows, err := q.db.QueryContext(ctx, fileQuery.String(), queryArgs...)
	if err != nil {
		log.Println(err)
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	result := agg.WithCount[agg.FileWithUser]{
		Total: totalFiles,
		Items: make([]agg.FileWithUser, 0),
	}

	for rows.Next() {
		var row agg.FileWithUser
		err := rows.Scan(
			&row.Id,
			&row.ChannelId,
			&row.UserId,
			&row.UserFirstName,
			&row.UserLastName,
			&row.OriginalName,
			&row.SavedName,
			&row.UploadedAt,
			&row.Url,
			&row.Size,
		)

		if err != nil {
			return nil, store.ErrQueryFailed
		}

		result.Items = append(result.Items, row)
	}

	return &result, nil
}
