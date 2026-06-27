package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *MessageRepository) Add(ctx context.Context, message *message.Message) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO messages(
            id,
            channel_id,
            sender_id,
            is_system,
            content,
            created_at
        ) VALUES (?, ?, ?, ?, ?, ?)`,
		message.ID().String(),
		message.ChannelID().String(),
		message.SenderID().String(),
		message.IsSystemMessage(),
		message.Content().String(),
		message.CreatedAt(),
	)
	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf(
					"message.MessageRepository.Add: duplicate message id: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"message.MessageRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *MessageRepository) Get(ctx context.Context, messageID message.MessageID) (*message.Message, error) {
	ex := r.executor(ctx)

	row := ex.QueryRowContext(
		ctx,
		`SELECT
            id,
            channel_id,
            sender_id,
            is_system,
            content,
            created_at
        FROM messages
        WHERE id = ?`,
		messageID.String(),
	)

	var (
		idRaw        string
		channelIDRaw string
		senderIDRaw  string
		isSystem     bool
		contentRaw   string
		createdAt    time.Time
	)

	err := row.Scan(
		&idRaw,
		&channelIDRaw,
		&senderIDRaw,
		&isSystem,
		&contentRaw,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", message.ErrMessageNotFound, err)
		}

		return nil, fmt.Errorf(
			"message.MessageRepository.Get: %w",
			err,
		)
	}

	id, err := message.NewMessageID(idRaw)
	if err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.Get: invalid message id: %w",
			err,
		)
	}

	channelID, err := message.NewChannelID(channelIDRaw)
	if err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.Get: invalid channel id: %w",
			err,
		)
	}

	senderID, err := iam.NewUserID(senderIDRaw)
	if err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.Get: invalid sender id: %w",
			err,
		)
	}

	content, err := message.NewContent(contentRaw)
	if err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.Get: invalid content: %w",
			err,
		)
	}

	return message.RestoreMessage(
		id,
		channelID,
		senderID,
		isSystem,
		content,
		createdAt,
	), nil
}

func (r *MessageRepository) ListByChannel(
	ctx context.Context,
	channelID message.ChannelID,
	before *message.MessageID,
	limit int,
) ([]*message.Message, error) {

	ex := r.executor(ctx)

	var query strings.Builder

	query.WriteString(`
    SELECT
        id,
        channel_id,
        sender_id,
        is_system,
        content,
        created_at
    FROM messages
    WHERE channel_id = ?
	`)

	args := []any{channelID.String()}

	if before != nil {
		query.WriteString(`
        AND id < ?
    `)
		args = append(args, before.String())
	}

	query.WriteString(`
    ORDER BY id DESC
    LIMIT ?
	`)
	args = append(args, limit)

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.ListByChannel: %w",
			err,
		)
	}

	defer rows.Close()

	messages := make([]*message.Message, 0)

	for rows.Next() {
		var (
			idRaw        string
			channelIDRaw string
			senderIDRaw  string
			isSystem     bool
			contentRaw   string
			createdAt    time.Time
		)

		err := rows.Scan(
			&idRaw,
			&channelIDRaw,
			&senderIDRaw,
			&isSystem,
			&contentRaw,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"message.MessageRepository.ListByChannel: %w",
				err,
			)
		}

		id, err := message.NewMessageID(idRaw)
		if err != nil {
			return nil, fmt.Errorf(
				"message.MessageRepository.ListByChannel: invalid message id: %w",
				err,
			)
		}

		parsedChannelID, err := message.NewChannelID(channelIDRaw)
		if err != nil {
			return nil, fmt.Errorf(
				"message.MessageRepository.ListByChannel: invalid channel id: %w",
				err,
			)
		}

		senderID, err := iam.NewUserID(senderIDRaw)
		if err != nil {
			return nil, fmt.Errorf(
				"message.MessageRepository.ListByChannel: invalid sender id: %w",
				err,
			)
		}

		content, err := message.NewContent(contentRaw)
		if err != nil {
			return nil, fmt.Errorf(
				"message.MessageRepository.ListByChannel: invalid content: %w",
				err,
			)
		}

		messages = append(
			messages,
			message.RestoreMessage(
				id,
				parsedChannelID,
				senderID,
				isSystem,
				content,
				createdAt,
			),
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.MessageRepository.ListByChannel: %w",
			err,
		)
	}

	return messages, nil
}
