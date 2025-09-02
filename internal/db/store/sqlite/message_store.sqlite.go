package sqlite

import (
	"context"
	"database/sql"
	"log"
	agg "mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"

	"github.com/mattn/go-sqlite3"
)

// SQLite implementation of MessageStore interface
type MessageStoreSqlite struct {
	db *sql.DB
}

// Creates a new instance of a MessageStoreSqlite
//
// Parameters:
//   - db: *sql.DB
//
// Returns:
//   - *MessageStoreSqlite
func NewMessageStore(db *sql.DB) *MessageStoreSqlite {
	return &MessageStoreSqlite{db: db}
}

// Implementation of CreateOne defined in MessageStore interface
func (q *MessageStoreSqlite) CreateOne(
	ctx context.Context,
	arg *params.MessageCreate) (*agg.MessageWithUser, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertMsgQuery := `
	INSERT INTO messages(channel_id, user_id, message, type) 
	VALUES(?, ?, ?, (SELECT id FROM message_types WHERE name = ?)) RETURNING id, created_at
	`

	var messageWithUser agg.MessageWithUser
	messageWithUser.UserId = arg.UserId
	messageWithUser.Message = arg.Message

	err = tx.QueryRowContext(ctx, insertMsgQuery,
		arg.ChannelId,
		arg.UserId,
		arg.Message,
		arg.Type).Scan(&messageWithUser.MessageId, &messageWithUser.CreatedAt)

	if err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return nil, store.ErrNotNullViolation
			}
		}

		return nil, store.ErrInsertFailed
	}

	userQuery := `
	SELECT u.first_name, u.last_name, u.title, u.image, r.name FROM users u
	JOIN roles r ON u.role = r.id
	WHERE u.id = ?
	`

	err = tx.QueryRowContext(ctx, userQuery, arg.UserId).Scan(
		&messageWithUser.FirstName,
		&messageWithUser.LastName,
		&messageWithUser.Title,
		&messageWithUser.Image,
		&messageWithUser.Role)
	if err != nil {
		log.Println(err)
		return nil, store.ErrQueryFailed
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
		return nil, store.ErrInsertFailed
	}

	return &messageWithUser, nil
}

// Implementation of GetChannelsByUserId defined in MessageStore interface
func (q *MessageStoreSqlite) GetChannelsByUserId(ctx context.Context, userId int64) ([]int64, error) {

	query := `SELECT channel_id FROM channel_users WHERE user_id = ?`

	channels := []int64{}
	rows, err := q.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	for rows.Next() {
		var id int64

		err := rows.Scan(&id)
		if err != nil {
			return nil, store.ErrQueryFailed
		}

		channels = append(channels, id)
	}

	return channels, nil
}

// Implementation of GetUsersByChannelId defined in MessageStore interface
func (q *MessageStoreSqlite) GetUsersByChannelId(ctx context.Context, channelId int64) ([]int64, error) {

	query := `SELECT user_id FROM channel_users WHERE channel_id = ?`

	users := []int64{}
	rows, err := q.db.QueryContext(ctx, query, channelId)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	for rows.Next() {
		var id int64

		err := rows.Scan(&id)
		if err != nil {
			return nil, store.ErrQueryFailed
		}

		users = append(users, id)
	}

	return users, nil
}

// GetByChannelId is an implementation of the GetByChannelId function
// defined in MessageStore interface. It returns an array of
// aggregates.MessageWithUser structs.
// Returns store.ErrQueryFailed if an error occurs.
func (q *MessageStoreSqlite) GetByChannelId(
	ctx context.Context,
	channelId int64,
	arg *params.MessageSearch) ([]agg.MessageWithUser, error) {

	messageQuery := `
	SELECT m.id, m.message, m.user_id, m.created_at, t.name, u.first_name, 
	u.last_name, u.title, u.image, r.name FROM messages m
	LEFT JOIN message_types t ON m.type = t.id
	LEFT JOIN users u ON m.user_id = u.id
	LEFT JOIN roles r ON u.role = r.id
	WHERE m.channel_id = ?
	LIMIT ? OFFSET ?
	`

	rows, err := q.db.QueryContext(ctx, messageQuery, channelId, arg.Limit, arg.Offset)
	if err != nil {
		return nil, store.ErrQueryFailed
	}

	defer rows.Close()

	var msgs []agg.MessageWithUser
	for rows.Next() {
		var msg agg.MessageWithUser
		rows.Scan(
			&msg.MessageId,
			&msg.Message,
			&msg.UserId,
			&msg.CreatedAt,
			&msg.Type,
			&msg.FirstName,
			&msg.LastName,
			&msg.Title,
			&msg.Image,
			&msg.Role,
		)

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// CountByChannelId is an implementation of the CountByChannelId function
// defined in MessageStore interface. It returns the number of messages
// in the specified channel.
// Returns store.ErrQueryFailed if an error occurs.
func (q *MessageStoreSqlite) CountByChannelId(ctx context.Context, channelId int64) (int64, error) {

	query := `SELECT COUNT(id) FROM messages WHERE channel_id = ?`

	var count int64
	if err := q.db.QueryRowContext(ctx, query, channelId).Scan(&count); err != nil {
		return 0, store.ErrQueryFailed
	}

	return count, nil
}
