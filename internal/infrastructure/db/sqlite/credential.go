package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"

	"github.com/mattn/go-sqlite3"
)

type CredentialRepository struct {
	db *sql.DB
}

func NewCredentialRepository(db *sql.DB) *CredentialRepository {
	return &CredentialRepository{
		db: db,
	}
}

func (r *CredentialRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *CredentialRepository) Add(ctx context.Context, c *iam.Credential) error {
	ex := r.executor(ctx)

	query := `
	INSERT INTO credentials (
		user_id,
		hash,
		version
	)
	VALUES (?, ?, ?)
	`

	_, err := ex.ExecContext(
		ctx,
		query,
		c.UserID().String(),
		c.Hash(),
		c.Version(),
	)
	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			switch sqlite3Err.ExtendedCode {
			case sqlite3.ErrConstraintPrimaryKey:
				return fmt.Errorf("iam.CredentialRepository.Add: duplicate credential: %w", err)
			case sqlite3.ErrConstraintForeignKey:
				return fmt.Errorf("iam.CredentialRepository.Add: user does not exist: %w", err)
			}
		}

		return fmt.Errorf("iam.CredentialRepository.Add: %w", err)
	}

	return nil
}

func (r *CredentialRepository) GetByUser(ctx context.Context, uID iam.UserID) (*iam.Credential, error) {
	ex := r.executor(ctx)

	const query = `
		SELECT
			user_id,
			hash,
			version
		FROM credentials
		WHERE user_id = ?
	`

	row := ex.QueryRowContext(ctx, query, uID.String())

	var (
		userIDRaw string
		hash      string
		version   int
	)

	err := row.Scan(&userIDRaw, &hash, &version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("iam.CredentialRepository.GetByUser: %w", iam.ErrCredentialNotFound)
		}
		return nil, fmt.Errorf("iam.CredentialRepository.GetByUser: %w", err)
	}

	userID, err := iam.NewUserID(userIDRaw)
	if err != nil {
		return nil, fmt.Errorf("iam.CredentialRepository.GetByUser: %w", err)
	}

	c := iam.RestoreCredential(
		userID,
		hash,
		version,
	)

	return c, nil
}

func (r *CredentialRepository) Save(ctx context.Context, c *iam.Credential) error {
	ex := r.executor(ctx)

	const query = `
		UPDATE credentials
		SET
			hash = ?,
			version = version + 1
		WHERE user_id = ?
		  AND version = ?
	`

	result, err := ex.ExecContext(
		ctx,
		query,
		c.Hash(),
		c.UserID().String(),
		c.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.CredentialRepository.Save: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.CredentialRepository.Save: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"%w: optimistic lock failed for credential %s",
			iam.ErrCredentialConcurrentModification,
			c.UserID(),
		)
	}

	return nil
}
