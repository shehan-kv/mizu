package sqlite

import (
	"context"
	"database/sql"
	"mizu/internal/db/models/aggregates"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
)

// ContractStoreSqlite implements the ContractStore interface using SQLite.
// It persists Contract entities in a SQLite database via the provided *sql.DB.
// The db connection must be non-nil, open, and initialized with the required
// contracts schema. Methods wrap lower-level SQL errors into domain-specific
// errors defined in the store package.
type ContractStoreSqlite struct {
	db *sql.DB
}

// NewContractStore constructs a ContractStoreSqlite that persists
// Contract entities in a SQLite database via the provided *sql.DB.
func NewContractStore(db *sql.DB) *ContractStoreSqlite {
	return &ContractStoreSqlite{db: db}
}

// CreateOne is an implementation of CreateOne in store.ContractStore.
// This method creates a new contract record, versions entry and
// system generated messages in relevant channels.
// It converts low-level driver-specific errors to domain specific
// errors defined in the store package.
//
//   - If a foreign key constraint violation occurs, it returns store.ErrForeignKeyViolation.
//   - If a unique constraint violation occurs, it returns store.ErrUniqueViolation.
//   - If a not-null constraint violation occurs, it returns store.ErrNotNullViolation.
//   - If any other errors occur, it returns store.ErrInsertFailed
func (q *ContractStoreSqlite) CreateOne(
	ctx context.Context,
	projectId int64,
	arg *params.ContractCreate) (*aggregates.ContractCreateResult, error) {

	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer tx.Rollback()

	insertContractQuery := `INSERT INTO contracts(project_id, name) VALUES(?, ?) RETURNING id`

	var contractId int64
	if err := tx.QueryRowContext(ctx, insertContractQuery, projectId, arg.Name).Scan(&contractId); err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return nil, store.ErrNotNullViolation
			}

			return nil, store.ErrInsertFailed
		}
	}

	insertVersionQuery := `
	INSERT INTO contract_versions(contract_id, version, contract)
	VALUES(?, ?, ?)
	`

	if _, err := tx.ExecContext(ctx, insertVersionQuery, contractId, arg.Version, arg.Contract); err != nil {
		if sqlite3Err, ok := err.(sqlite3.Error); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.ErrForeignKeyViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return nil, store.ErrUniqueViolation
			}

			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintNotNull {
				return nil, store.ErrNotNullViolation
			}

			return nil, store.ErrInsertFailed
		}
	}

	channelsQuery := `SELECT id FROM channels WHERE project_id = ?`

	rows, err := tx.QueryContext(ctx, channelsQuery, projectId)
	if err != nil {
		return nil, store.ErrInsertFailed
	}

	defer rows.Close()

	var channelIds []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, store.ErrInsertFailed
		}

		channelIds = append(channelIds, id)
	}

	result := aggregates.ContractCreateResult{
		ContractId: contractId,
		Version:    arg.Version,
		Messages:   make(map[int64]aggregates.MessageWithUser),
		UserIds:    make(map[int64][]int64),
	}

	userIdsQuery := `SELECT DISTINCT(user_id) FROM channel_users WHERE channel_id = ?`

	for _, channelId := range channelIds {
		usrRows, err := tx.QueryContext(ctx, userIdsQuery, channelId)
		if err != nil {
			return nil, store.ErrInsertFailed
		}

		var userIds []int64
		for usrRows.Next() {
			var id int64
			if err := usrRows.Scan(&id); err != nil {
				if err := usrRows.Close(); err != nil {
					return nil, store.ErrInsertFailed
				}
				return nil, store.ErrInsertFailed
			}

			userIds = append(userIds, id)
		}

		result.UserIds[channelId] = userIds

		if err := usrRows.Close(); err != nil {
			return nil, store.ErrInsertFailed
		}
	}

	insertMsgQuery := `
	INSERT INTO messages(user_id, channel_id, message, type) 
	VALUES(?, ?, ?, (SELECT id FROM message_types WHERE name = ?)) RETURNING id, created_at
	`

	for _, channelId := range channelIds {

		var msgId int64
		var createdAt time.Time

		err := tx.QueryRowContext(
			ctx,
			insertMsgQuery,
			nil, channelId,
			strconv.FormatInt(contractId, 10),
			params.MessageTypeContract,
		).Scan(&msgId, &createdAt)

		if err != nil {
			return nil, store.ErrInsertFailed
		}

		msg := aggregates.MessageWithUser{
			UserId:    0,
			MessageId: msgId,
			ChannelId: channelId,
			Message:   strconv.FormatInt(contractId, 10),
			CreatedAt: createdAt,
			FirstName: "",
			LastName:  "",
			Type:      params.MessageTypeContract,
			Role:      "",
			Title:     "",
			Image:     nil,
		}

		result.Messages[channelId] = msg
	}

	if err = tx.Commit(); err != nil {
		return nil, store.ErrInsertFailed
	}

	return &result, nil
}
