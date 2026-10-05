package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"strconv"
	"strings"
	"time"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{
		db: db,
	}
}

func (r *FileRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *FileRepository) Add(ctx context.Context, f *message.File) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO files(
			id,
			channel_id,
			user_id,
			original_name,
			saved_name,
			storage_key,
			mime_type,
			size,
			uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		f.ID().String(),
		f.ChannelID().String(),
		f.UserID().String(),
		f.OriginalName().String(),
		f.SavedName().String(),
		f.StorageKey(),
		f.MimeType(),
		f.Size(),
		f.UploadedAt(),
	)

	if err != nil {
		return fmt.Errorf("message.FileRepository.Add: %w", err)
	}

	return nil
}

func (r *FileRepository) Get(ctx context.Context, id message.FileID) (*message.File, error) {
	ex := r.executor(ctx)

	row := ex.QueryRowContext(
		ctx,
		`SELECT
			id,
			channel_id,
			user_id,
			original_name,
			saved_name,
			storage_key,
			mime_type,
			size,
			uploaded_at
		FROM files
		WHERE id = $1`,
		id.String(),
	)

	var (
		fileIDValue       string
		channelIDValue    string
		userIDValue       string
		originalNameValue string
		savedNameValue    string
		storageKey        string
		mimeType          string
		size              int64
		uploadedAt        time.Time
	)

	err := row.Scan(
		&fileIDValue,
		&channelIDValue,
		&userIDValue,
		&originalNameValue,
		&savedNameValue,
		&storageKey,
		&mimeType,
		&size,
		&uploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(
				"message.FileRepository.Get: file not found: %w",
				err,
			)
		}

		return nil, fmt.Errorf(
			"message.FileRepository.Get: %w",
			err,
		)
	}

	fileID, err := message.NewFileID(fileIDValue)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.Get: invalid file id: %w",
			err,
		)
	}

	channelID, err := message.NewChannelID(channelIDValue)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.Get: invalid channel id: %w",
			err,
		)
	}

	userID, err := iam.NewUserID(userIDValue)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.Get: invalid user id: %w",
			err,
		)
	}

	originalName, err := message.NewFileName(originalNameValue)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.Get: invalid original file name: %w",
			err,
		)
	}

	savedName, err := message.NewFileName(savedNameValue)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.Get: invalid saved file name: %w",
			err,
		)
	}

	file := message.RestoreFile(
		fileID,
		channelID,
		userID,
		originalName,
		savedName,
		storageKey,
		mimeType,
		size,
		uploadedAt,
	)

	return file, nil
}

func (r *FileRepository) GetStatsByProject(ctx context.Context, pID project.ProjectID) (message.Stats, error) {
	ex := r.executor(ctx)

	row := ex.QueryRowContext(
		ctx,
		`SELECT COUNT(f.id)
		FROM files f
		INNER JOIN channels c
			ON c.id = f.channel_id
		WHERE c.project_id = $1`,
		pID.String(),
	)

	var fileCount int

	err := row.Scan(&fileCount)
	if err != nil {
		return message.Stats{}, fmt.Errorf(
			"message.FileRepository.GetStatsByProject: %w",
			err,
		)
	}

	stats := message.NewStats(fileCount)

	return stats, nil
}

func (r *FileRepository) ListByChannel(ctx context.Context, f message.ChannelFileFilter, p common.Page) ([]*message.File, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 6)

	argPos := 1

	query.WriteString(`
        SELECT
            id,
            channel_id,
            user_id,
            original_name,
            saved_name,
            storage_key,
            mime_type,
            size,
            uploaded_at
        FROM files
        WHERE channel_id = $1
    `)

	args = append(args, f.ChannelID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(` AND (original_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(` OR saved_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(`)`)

		keyword := "%" + *f.Keyword + "%"
		args = append(args, keyword, keyword)
		argPos += 2
	}

	query.WriteString(` ORDER BY uploaded_at DESC LIMIT $`)
	query.WriteString(strconv.Itoa(argPos))
	query.WriteString(` OFFSET $`)
	query.WriteString(strconv.Itoa(argPos + 1))

	args = append(args, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByChannel: %w",
			err,
		)
	}
	defer rows.Close()

	files := make([]*message.File, 0)

	for rows.Next() {
		var (
			fileIDValue       string
			channelIDValue    string
			userIDValue       string
			originalNameValue string
			savedNameValue    string
			storageKey        string
			mimeType          string
			size              int64
			uploadedAt        time.Time
		)

		err = rows.Scan(
			&fileIDValue,
			&channelIDValue,
			&userIDValue,
			&originalNameValue,
			&savedNameValue,
			&storageKey,
			&mimeType,
			&size,
			&uploadedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: %w",
				err,
			)
		}

		fileID, err := message.NewFileID(fileIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByChannel: invalid file id: %w", err)
		}

		channelID, err := message.NewChannelID(channelIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByChannel: invalid channel id: %w", err)
		}

		userID, err := iam.NewUserID(userIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByChannel: invalid user id: %w", err)
		}

		originalName, err := message.NewFileName(originalNameValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByChannel: invalid original file name: %w", err)
		}

		savedName, err := message.NewFileName(savedNameValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByChannel: invalid saved file name: %w", err)
		}

		file := message.RestoreFile(
			fileID,
			channelID,
			userID,
			originalName,
			savedName,
			storageKey,
			mimeType,
			size,
			uploadedAt,
		)

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByChannel: %w",
			err,
		)
	}

	return files, nil
}

func (r *FileRepository) ListByProject(ctx context.Context, f message.ProjectFileFilter, p common.Page) ([]*message.File, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 6)

	query.WriteString(`
	SELECT DISTINCT
		f.id,
		f.channel_id,
		f.user_id,
		f.original_name,
		f.saved_name,
		f.storage_key,
		f.mime_type,
		f.size,
		f.uploaded_at
	FROM files f
    INNER JOIN channels c ON c.id = f.channel_id
    INNER JOIN channel_members cm ON cm.channel_id = c.id
    WHERE c.project_id = $1
      AND cm.user_id = $2
    `)

	argPos := 2

	args = append(args, f.ProjectID.String(), f.MemberID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(` AND (f.original_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(` OR f.saved_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(`)`)

		keyword := "%" + *f.Keyword + "%"
		args = append(args, keyword, keyword)
		argPos += 2
	}

	query.WriteString(` ORDER BY f.uploaded_at DESC LIMIT $`)
	query.WriteString(strconv.Itoa(argPos))
	query.WriteString(` OFFSET $`)
	query.WriteString(strconv.Itoa(argPos + 1))

	args = append(args, p.Limit(), p.Offset())

	rows, err := ex.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByProject: %w",
			err,
		)
	}
	defer rows.Close()

	files := make([]*message.File, 0)

	for rows.Next() {
		var (
			fileIDValue       string
			channelIDValue    string
			userIDValue       string
			originalNameValue string
			savedNameValue    string
			storageKey        string
			mimeType          string
			size              int64
			uploadedAt        time.Time
		)

		err = rows.Scan(
			&fileIDValue,
			&channelIDValue,
			&userIDValue,
			&originalNameValue,
			&savedNameValue,
			&storageKey,
			&mimeType,
			&size,
			&uploadedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: %w",
				err,
			)
		}

		fileID, err := message.NewFileID(fileIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByProject: invalid file id: %w", err)
		}

		channelID, err := message.NewChannelID(channelIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByProject: invalid channel id: %w", err)
		}

		userID, err := iam.NewUserID(userIDValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByProject: invalid user id: %w", err)
		}

		originalName, err := message.NewFileName(originalNameValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByProject: invalid original file name: %w", err)
		}

		savedName, err := message.NewFileName(savedNameValue)
		if err != nil {
			return nil, fmt.Errorf("message.FileRepository.ListByProject: invalid saved file name: %w", err)
		}

		file := message.RestoreFile(
			fileID,
			channelID,
			userID,
			originalName,
			savedName,
			storageKey,
			mimeType,
			size,
			uploadedAt,
		)

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByProject: %w",
			err,
		)
	}

	return files, nil
}

func (r *FileRepository) CountByChannel(ctx context.Context, f message.ChannelFileFilter) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 4)

	argPos := 1

	query.WriteString(`
        SELECT COUNT(*)
        FROM files
        WHERE channel_id = $1
    `)

	args = append(args, f.ChannelID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(` AND (original_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(` OR saved_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(`)`)

		keyword := "%" + *f.Keyword + "%"
		args = append(args, keyword, keyword)
		argPos += 2
	}

	row := ex.QueryRowContext(ctx, query.String(), args...)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"message.FileRepository.CountByChannel: %w",
			err,
		)
	}

	return count, nil
}

func (r *FileRepository) CountByProject(ctx context.Context, f message.ProjectFileFilter) (int, error) {
	ex := r.executor(ctx)

	var query strings.Builder
	args := make([]any, 0, 4)

	argPos := 1

	query.WriteString(`
        SELECT COUNT(*)
        FROM files f
        INNER JOIN channels c ON c.id = f.channel_id
        WHERE c.project_id = $1
    `)

	args = append(args, f.ProjectID.String())
	argPos++

	if f.Keyword != nil {
		query.WriteString(` AND (f.original_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos))
		query.WriteString(` OR f.saved_name LIKE $`)
		query.WriteString(strconv.Itoa(argPos + 1))
		query.WriteString(`)`)

		keyword := "%" + *f.Keyword + "%"
		args = append(args, keyword, keyword)
		argPos += 2
	}

	row := ex.QueryRowContext(ctx, query.String(), args...)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"message.FileRepository.CountByProject: %w",
			err,
		)
	}

	return count, nil
}
