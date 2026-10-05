package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"
	"time"

	"github.com/mattn/go-sqlite3"
)

type RecoveryRepository struct {
	db *sql.DB
}

func NewRecoveryRepository(db *sql.DB) *RecoveryRepository {
	return &RecoveryRepository{
		db: db,
	}
}

func (r *RecoveryRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *RecoveryRepository) Add(ctx context.Context, recovery *iam.Recovery) error {
	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO recoveries(
			user_id,
			token,
			version,
			created_at
		) VALUES (?, ?, ?, ?)`,
		recovery.UserID().String(),
		recovery.Token().String(),
		recovery.Version(),
		recovery.CreatedAt(),
	)
	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			switch sqlite3Err.ExtendedCode {
			case sqlite3.ErrConstraintPrimaryKey:
				return fmt.Errorf(
					"iam.RecoveryRepository.Add: duplicate recovery: %w",
					err,
				)

			case sqlite3.ErrConstraintForeignKey:
				return fmt.Errorf(
					"iam.RecoveryRepository.Add: foreign key constraint violation: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"iam.RecoveryRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *RecoveryRepository) GetByToken(ctx context.Context, token iam.RecoveryToken) (*iam.Recovery, error) {
	ex := r.executor(ctx)

	var (
		rawUserID string
		rawToken  string
		version   int
		createdAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			user_id,
			token,
			version,
			created_at
		FROM recoveries
		WHERE token = ?`,
		token.String(),
	).Scan(
		&rawUserID,
		&rawToken,
		&version,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrRecoveryNotFound, err)
		}

		return nil, fmt.Errorf(
			"iam.RecoveryRepository.GetByToken: %w",
			err,
		)
	}

	userID, err := iam.NewUserID(rawUserID)
	if err != nil {
		return nil, err
	}

	recoveryToken, err := iam.NewRecoveryToken(rawToken)
	if err != nil {
		return nil, err
	}

	return iam.RestoreRecovery(
		userID,
		recoveryToken,
		version,
		createdAt,
	), nil
}

func (r *RecoveryRepository) GetByUser(ctx context.Context, uID iam.UserID) (*iam.Recovery, error) {
	ex := r.executor(ctx)

	var (
		rawUserID string
		rawToken  string
		version   int
		createdAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			user_id,
			token,
			version,
			created_at
		FROM recoveries
		WHERE user_id = ?`,
		uID.String(),
	).Scan(
		&rawUserID,
		&rawToken,
		&version,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrRecoveryNotFound, err)
		}

		return nil, fmt.Errorf(
			"iam.RecoveryRepository.GetByUser: %w",
			err,
		)
	}

	userID, err := iam.NewUserID(rawUserID)
	if err != nil {
		return nil, err
	}

	recoveryToken, err := iam.NewRecoveryToken(rawToken)
	if err != nil {
		return nil, err
	}

	return iam.RestoreRecovery(
		userID,
		recoveryToken,
		version,
		createdAt,
	), nil
}

func (r *RecoveryRepository) Save(ctx context.Context, recovery *iam.Recovery) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`UPDATE recoveries
		SET
			token = ?,
			version = ?,
			created_at = ?
		WHERE
			user_id = ?
			AND version = ?`,
		recovery.Token().String(),
		recovery.Version()+1,
		recovery.CreatedAt(),
		recovery.UserID().String(),
		recovery.Version(),
	)
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.Save: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.Save: rows affected: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"iam.RecoveryRepository.Save: %w",
			iam.ErrRecoveryConcurrentModification,
		)
	}

	return nil
}

func (r *RecoveryRepository) Remove(ctx context.Context, recovery *iam.Recovery) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`DELETE FROM recoveries
		WHERE user_id = ? AND version = ?`,
		recovery.UserID().String(),
		recovery.Version(),
	)
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.Remove: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.Remove: rows affected: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"iam.RecoveryRepository.Remove: %w",
			iam.ErrRecoveryConcurrentModification,
		)
	}

	return nil
}

func (r *RecoveryRepository) RemoveByUser(ctx context.Context, uID iam.UserID) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`DELETE FROM recoveries
		WHERE user_id = ?`,
		uID.String(),
	)
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.RemoveByUser: %w",
			err,
		)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"iam.RecoveryRepository.RemoveByUser: rows affected: %w",
			err,
		)
	}

	return nil
}
