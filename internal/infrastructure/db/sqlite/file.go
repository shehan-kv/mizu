package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/message"
	"mizu/internal/domain/project"
	"time"

	"github.com/mattn/go-sqlite3"
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
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf(
					"message.FileRepository.Add: duplicate file id: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"message.FileRepository.Add: %w",
			err,
		)
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
        WHERE id = ?`,
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
        WHERE c.project_id = ?`,
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

	stats := message.NewStats(
		fileCount,
	)

	return stats, nil
}

func (r *FileRepository) ListByChannel(ctx context.Context, f message.ChannelFileFilter, p common.Page) ([]*message.File, error) {
	ex := r.executor(ctx)

	query := `
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
        WHERE channel_id = ?
    `

	args := make([]any, 0, 5)
	args = append(args, f.ChannelID.String())

	if f.Keyword != nil {
		query += `
            AND (
                original_name LIKE ?
                OR saved_name LIKE ?
            )
        `

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	query += `
        ORDER BY uploaded_at DESC
        LIMIT ?
        OFFSET ?
    `

	args = append(
		args,
		p.Limit(),
		p.Offset(),
	)

	rows, err := ex.QueryContext(
		ctx,
		query,
		args...,
	)
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
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: invalid file id: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(channelIDValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: invalid channel id: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(userIDValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: invalid user id: %w",
				err,
			)
		}

		originalName, err := message.NewFileName(originalNameValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: invalid original file name: %w",
				err,
			)
		}

		savedName, err := message.NewFileName(savedNameValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByChannel: invalid saved file name: %w",
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

		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByChannel: %w",
			err,
		)
	}

	return files, nil
}

func (r *FileRepository) ListByProject(ctx context.Context, f message.ProjectFileFilter, p common.Page) ([]*message.File, error) {
	ex := r.executor(ctx)

	query := `
        SELECT
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
		WHERE c.project_id = ?
    `

	args := make([]any, 0, 5)
	args = append(args, f.ProjectID.String())

	if f.Keyword != nil {
		query += `
            AND (
                original_name LIKE ?
                OR saved_name LIKE ?
            )
        `

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	query += `
        ORDER BY uploaded_at DESC
        LIMIT ?
        OFFSET ?
    `

	args = append(
		args,
		p.Limit(),
		p.Offset(),
	)

	rows, err := ex.QueryContext(
		ctx,
		query,
		args...,
	)
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
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: invalid file id: %w",
				err,
			)
		}

		channelID, err := message.NewChannelID(channelIDValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: invalid channel id: %w",
				err,
			)
		}

		userID, err := iam.NewUserID(userIDValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: invalid user id: %w",
				err,
			)
		}

		originalName, err := message.NewFileName(originalNameValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: invalid original file name: %w",
				err,
			)
		}

		savedName, err := message.NewFileName(savedNameValue)
		if err != nil {
			return nil, fmt.Errorf(
				"message.FileRepository.ListByProject: invalid saved file name: %w",
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

		files = append(files, file)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf(
			"message.FileRepository.ListByProject: %w",
			err,
		)
	}

	return files, nil
}

func (r *FileRepository) CountByChannel(ctx context.Context, f message.ChannelFileFilter) (int, error) {
	ex := r.executor(ctx)

	query := `
        SELECT COUNT(*)
        FROM files
        WHERE channel_id = ?
    `

	args := make([]any, 0, 3)
	args = append(args, f.ChannelID.String())

	if f.Keyword != nil {
		query += `
            AND (
                original_name LIKE ?
                OR saved_name LIKE ?
            )
        `

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	row := ex.QueryRowContext(
		ctx,
		query,
		args...,
	)

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

	query := `
        SELECT COUNT(*)
        FROM files f
        INNER JOIN channels c ON c.id = f.channel_id
		WHERE c.project_id = ?
    `

	args := make([]any, 0, 3)
	args = append(args, f.ProjectID.String())

	if f.Keyword != nil {
		query += `
            AND (
                original_name LIKE ?
                OR saved_name LIKE ?
            )
        `

		keyword := "%" + *f.Keyword + "%"

		args = append(
			args,
			keyword,
			keyword,
		)
	}

	row := ex.QueryRowContext(
		ctx,
		query,
		args...,
	)

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
