package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
	"time"

	"github.com/mattn/go-sqlite3"
)

type VerificationRepository struct {
	db *sql.DB
}

func NewVerificationRepository(db *sql.DB) *VerificationRepository {
	return &VerificationRepository{
		db: db,
	}
}

func (r *VerificationRepository) executor(ctx context.Context) executor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func (r *VerificationRepository) Add(ctx context.Context, v *verification.Verification) error {

	ex := r.executor(ctx)

	_, err := ex.ExecContext(
		ctx,
		`INSERT INTO verifications(
            id,
            user_id,
            version,
            created_at
        ) VALUES (?, ?, ?, ?)`,
		v.ID().String(),
		v.UserID().String(),
		v.Version(),
		v.CreatedAt(),
	)
	if err != nil {
		if sqlite3Err, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqlite3Err.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return fmt.Errorf(
					"verification.VerificationRepository.Add: duplicate verification id: %w",
					err,
				)
			}
		}

		return fmt.Errorf(
			"verification.VerificationRepository.Add: %w",
			err,
		)
	}

	return nil
}

func (r *VerificationRepository) Get(ctx context.Context, id verification.VerificationID) (*verification.Verification, error) {
	ex := r.executor(ctx)

	var (
		rawID     string
		rawUserID string
		version   int
		createdAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
            id,
            user_id,
            version,
            created_at
        FROM verifications
        WHERE id = ?`,
		id.String(),
	).Scan(
		&rawID,
		&rawUserID,
		&version,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", verification.ErrVerificationNotFound, err)
		}

		return nil, fmt.Errorf(
			"verification.VerificationRepository.Get: %w",
			err,
		)
	}

	verificationID, err := verification.NewVerificationID(rawID)
	if err != nil {
		return nil, err
	}

	userID, err := iam.NewUserID(rawUserID)
	if err != nil {
		return nil, err
	}

	v := verification.RestoreVerification(
		verificationID,
		userID,
		version,
		createdAt,
	)

	return v, nil
}

func (r *VerificationRepository) Remove(ctx context.Context, v *verification.Verification) error {

	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`DELETE FROM verifications WHERE id = ? AND version = ?`,
		v.ID().String(),
		v.Version(),
	)
	if err != nil {
		return fmt.Errorf(
			"verification.VerificationRepository.Remove: %w",
			err,
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"verification.VerificationRepository.Remove: rows affected: %w",
			err,
		)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(
			"verification.VerificationRepository.Remove: %w", verification.ErrVerificationConcurrentModification,
		)
	}

	return nil
}

func (r *VerificationRepository) RemoveByUserID(ctx context.Context, uID iam.UserID) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`DELETE FROM verifications WHERE user_id = ?`,
		uID.String(),
	)
	if err != nil {
		return fmt.Errorf(
			"verification.VerificationRepository.RemoveByUserID: %w",
			err,
		)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"verification.VerificationRepository.RemoveByUserID: rows affected: %w",
			err,
		)
	}

	return nil
}
