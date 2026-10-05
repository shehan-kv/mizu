package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ChannelRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository {
	return &ChannelRepository{
		db: db,
	}
}

func (r *ChannelRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *ChannelRepository) Add(ctx context.Context, channel *message.Channel) error {
	ex := r.executor(ctx)

	var projectID any
	if channel.ProjectID() != nil {
		projectID = channel.ProjectID().String()
	}

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO channels (
			id,
			project_id,
			name,
			version,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		channel.ID().String(),
		projectID,
		channel.Name().String(),
		channel.Version(),
		channel.CreatedAt(),
		channel.UpdatedAt(),
	)
	if err != nil {

		if pqErr, ok := errors.AsType[*pq.Error](err); ok {
			if pqErr.Code == "23505" &&
				pqErr.Constraint == "channels_pkey" {
				return fmt.Errorf(
					"message.ChannelRepository.Add: duplicate channel id: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"message.ChannelRepository.Add: %w",
			err,
		)
	}

	members := channel.Members()

	if len(members) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteString(`
		INSERT INTO channel_members (
			channel_id,
			user_id
		) VALUES `)

	args := make([]any, 0, len(members)*2)

	for idx, memberID := range members {
		if idx > 0 {
			sb.WriteString(", ")
		}

		p1 := idx*2 + 1
		p2 := idx*2 + 2

		sb.WriteByte('(')
		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(p1))
		sb.WriteString(", $")
		sb.WriteString(strconv.Itoa(p2))
		sb.WriteByte(')')

		args = append(
			args,
			channel.ID().String(),
			memberID.String(),
		)
	}

	_, err = ex.ExecContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"message.ChannelRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *ChannelRepository) Get(ctx context.Context, channelID message.ChannelID) (*message.Channel, error) {
	ex := r.executor(ctx)

	var (
		id              string
		projectIDString *string
		name            string
		version         int
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			id,
			project_id,
			name,
			version,
			created_at,
			updated_at
		FROM channels
		WHERE id = $1`,
		channelID.String(),
	).Scan(
		&id,
		&projectIDString,
		&name,
		&version,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, message.ErrChannelNotFound
		}

		return nil, fmt.Errorf(
			"message.ChannelRepository.Get: %w",
			err,
		)
	}

	rows, err := ex.QueryContext(
		ctx,
		`SELECT user_id
		FROM channel_members
		WHERE channel_id = $1`,
		channelID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.Get: %w",
			err,
		)
	}
	defer rows.Close()

	members := make([]iam.UserID, 0)

	for rows.Next() {
		var userIDString string

		if err := rows.Scan(&userIDString); err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.Get: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(userIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.Get: invalid user id: %w",
				err,
			)
		}

		members = append(members, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.Get: %w",
			err,
		)
	}

	var projectID *project.ProjectID

	if projectIDString != nil {
		id, err := project.NewProjectID(*projectIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.Get: invalid project id: %w",
				err,
			)
		}

		projectID = &id
	}

	channelIDValue, err := message.NewChannelID(id)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.Get: invalid channel id: %w",
			err,
		)
	}

	channelName, err := message.NewChannelName(name)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.Get: invalid channel name: %w",
			err,
		)
	}

	channel := message.RestoreChannel(
		channelIDValue,
		projectID,
		channelName,
		members,
		version,
		createdAt,
		updatedAt,
	)

	return channel, nil
}

func (r *ChannelRepository) ListByProject(ctx context.Context, projectID project.ProjectID) ([]*message.Channel, error) {
	ex := r.executor(ctx)

	rows, err := ex.QueryContext(
		ctx,
		`SELECT
			id,
			project_id,
			name,
			version,
			created_at,
			updated_at
		FROM channels
		WHERE project_id = $1`,
		projectID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByProject: %w",
			err,
		)
	}
	defer rows.Close()

	type channelRow struct {
		id        message.ChannelID
		projectID *project.ProjectID
		name      message.ChannelName
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	channelRows := make([]channelRow, 0)
	channelIDs := make([]any, 0)

	for rows.Next() {
		var (
			idString        string
			projectIDString *string
			nameString      string
			version         int
			createdAt       time.Time
			updatedAt       time.Time
		)

		if err := rows.Scan(
			&idString,
			&projectIDString,
			&nameString,
			&version,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(idString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: invalid channel id: %w",
				err,
			)
		}

		name, err := message.NewChannelName(nameString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: invalid channel name: %w",
				err,
			)
		}

		var pid *project.ProjectID

		if projectIDString != nil {
			value, err := project.NewProjectID(*projectIDString)
			if err != nil {
				return nil, fmt.Errorf(
					"message.ChannelRepository.ListByProject: invalid project id: %w",
					err,
				)
			}

			pid = &value
		}

		channelRows = append(channelRows, channelRow{
			id:        channelID,
			projectID: pid,
			name:      name,
			version:   version,
			createdAt: createdAt,
			updatedAt: updatedAt,
		})

		channelIDs = append(channelIDs, channelID.String())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByProject: %w",
			err,
		)
	}

	if len(channelRows) == 0 {
		return []*message.Channel{}, nil
	}

	var sb strings.Builder

	sb.WriteString(`
		SELECT
			channel_id,
			user_id
		FROM channel_members
		WHERE channel_id IN (`)

	for i := range channelIDs {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(i + 1))
	}

	sb.WriteByte(')')

	memberRows, err := ex.QueryContext(
		ctx,
		sb.String(),
		channelIDs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByProject: %w",
			err,
		)
	}
	defer memberRows.Close()

	memberMap := make(map[message.ChannelID][]iam.UserID)

	for memberRows.Next() {
		var (
			channelIDString string
			userIDString    string
		)

		if err := memberRows.Scan(
			&channelIDString,
			&userIDString,
		); err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(channelIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: invalid channel id: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(userIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByProject: invalid user id: %w",
				err,
			)
		}

		memberMap[channelID] = append(
			memberMap[channelID],
			userID,
		)
	}

	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByProject: %w",
			err,
		)
	}

	channels := make([]*message.Channel, 0, len(channelRows))

	for i := range channelRows {
		row := channelRows[i]

		channels = append(
			channels,
			message.RestoreChannel(
				row.id,
				row.projectID,
				row.name,
				memberMap[row.id],
				row.version,
				row.createdAt,
				row.updatedAt,
			),
		)
	}

	return channels, nil
}

func (r *ChannelRepository) ListByMember(ctx context.Context, memberID iam.UserID) ([]*message.Channel, error) {
	ex := r.executor(ctx)

	rows, err := ex.QueryContext(
		ctx,
		`SELECT
			c.id,
			c.project_id,
			c.name,
			c.version,
			c.created_at,
			c.updated_at
		FROM channels c
		INNER JOIN channel_members cm
			ON cm.channel_id = c.id
		WHERE cm.user_id = $1`,
		memberID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByMember: %w",
			err,
		)
	}
	defer rows.Close()

	type channelRow struct {
		id        message.ChannelID
		projectID *project.ProjectID
		name      message.ChannelName
		version   int
		createdAt time.Time
		updatedAt time.Time
	}

	channelRows := make([]channelRow, 0)
	channelIDs := make([]any, 0)

	for rows.Next() {
		var (
			idString        string
			projectIDString *string
			nameString      string
			version         int
			createdAt       time.Time
			updatedAt       time.Time
		)

		if err := rows.Scan(
			&idString,
			&projectIDString,
			&nameString,
			&version,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(idString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: invalid channel id: %w",
				err,
			)
		}

		name, err := message.NewChannelName(nameString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: invalid channel name: %w",
				err,
			)
		}

		var pid *project.ProjectID

		if projectIDString != nil {
			value, err := project.NewProjectID(*projectIDString)
			if err != nil {
				return nil, fmt.Errorf(
					"message.ChannelRepository.ListByMember: invalid project id: %w",
					err,
				)
			}

			pid = &value
		}

		channelRows = append(channelRows, channelRow{
			id:        channelID,
			projectID: pid,
			name:      name,
			version:   version,
			createdAt: createdAt,
			updatedAt: updatedAt,
		})

		channelIDs = append(channelIDs, channelID.String())
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByMember: %w",
			err,
		)
	}

	if len(channelRows) == 0 {
		return []*message.Channel{}, nil
	}

	var sb strings.Builder

	sb.WriteString(`
		SELECT
			channel_id,
			user_id
		FROM channel_members
		WHERE channel_id IN (`)

	for i := range channelIDs {
		if i > 0 {
			sb.WriteString(", ")
		}

		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(i + 1))
	}

	sb.WriteByte(')')

	memberRows, err := ex.QueryContext(
		ctx,
		sb.String(),
		channelIDs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByMember: %w",
			err,
		)
	}
	defer memberRows.Close()

	memberMap := make(map[message.ChannelID][]iam.UserID)

	for memberRows.Next() {
		var (
			channelIDString string
			userIDString    string
		)

		if err := memberRows.Scan(
			&channelIDString,
			&userIDString,
		); err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(channelIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: invalid channel id: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(userIDString)
		if err != nil {
			return nil, fmt.Errorf(
				"message.ChannelRepository.ListByMember: invalid user id: %w",
				err,
			)
		}

		memberMap[channelID] = append(
			memberMap[channelID],
			userID,
		)
	}

	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.ChannelRepository.ListByMember: %w",
			err,
		)
	}

	channels := make([]*message.Channel, 0, len(channelRows))

	for i := range channelRows {
		row := channelRows[i]

		channels = append(
			channels,
			message.RestoreChannel(
				row.id,
				row.projectID,
				row.name,
				memberMap[row.id],
				row.version,
				row.createdAt,
				row.updatedAt,
			),
		)
	}

	return channels, nil
}

func (r *ChannelRepository) Save(ctx context.Context, channel *message.Channel) error {
	ex := r.executor(ctx)

	var projectID any
	if channel.ProjectID() != nil {
		projectID = channel.ProjectID().String()
	}

	result, err := ex.ExecContext(
		ctx,
		`UPDATE channels
		SET
			project_id = $1,
			name = $2,
			version = version + 1,
			updated_at = $3
		WHERE id = $4
			AND version = $5`,
		projectID,
		channel.Name().String(),
		channel.UpdatedAt(),
		channel.ID().String(),
		channel.Version(),
	)
	if err != nil {
		return fmt.Errorf(
			"message.ChannelRepository.Save: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"message.ChannelRepository.Save: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"%w: version mismatch",
			message.ErrChannelConcurrentModification,
		)
	}

	_, err = ex.ExecContext(
		ctx,
		`DELETE FROM channel_members
		WHERE channel_id = $1`,
		channel.ID().String(),
	)
	if err != nil {
		return fmt.Errorf(
			"message.ChannelRepository.Save: %w",
			err,
		)
	}

	members := channel.Members()

	var sb strings.Builder

	sb.WriteString(`
		INSERT INTO channel_members (
			channel_id,
			user_id
		) VALUES `)

	args := make([]any, 0, len(members)*2)

	for i := range members {
		if i > 0 {
			sb.WriteString(", ")
		}

		p1 := i*2 + 1
		p2 := i*2 + 2

		sb.WriteByte('(')
		sb.WriteByte('$')
		sb.WriteString(strconv.Itoa(p1))
		sb.WriteString(", $")
		sb.WriteString(strconv.Itoa(p2))
		sb.WriteByte(')')

		args = append(
			args,
			channel.ID().String(),
			members[i].String(),
		)
	}

	_, err = ex.ExecContext(
		ctx,
		sb.String(),
		args...,
	)
	if err != nil {
		return fmt.Errorf(
			"message.ChannelRepository.Save: %w",
			err,
		)
	}

	return nil
}
