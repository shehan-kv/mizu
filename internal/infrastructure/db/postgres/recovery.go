package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mizu/internal/domain/iam"
	"time"

	"github.com/lib/pq"
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
			created_at,
			expires_at
		) VALUES ($1, $2, $3, $4, $5)`,
		recovery.UserID().String(),
		recovery.Token().String(),
		recovery.Version(),
		recovery.CreatedAt(),
		recovery.ExpiresAt(),
	)
	if err != nil {
		if pqErr, ok := errors.AsType[*pq.Error](err); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf(
					"iam.RecoveryRepository.Add: duplicate recovery (unique constraint): %w",
					err,
				)

			case "23503": // foreign_key_violation
				return fmt.Errorf(
					"iam.RecoveryRepository.Add: foreign key constraint violation: %w",
					err,
				)
			}
		}

		return fmt.Errorf("iam.RecoveryRepository.Add: %w", err)
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
		expiresAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			user_id,
			token,
			version,
			created_at,
			expires_at
		FROM recoveries
		WHERE token = $1`,
		token.String(),
	).Scan(
		&rawUserID,
		&rawToken,
		&version,
		&createdAt,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrRecoveryNotFound, err)
		}

		return nil, fmt.Errorf("iam.RecoveryRepository.GetByToken: %w", err)
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
		expiresAt,
	), nil
}

func (r *RecoveryRepository) GetByUser(ctx context.Context, uID iam.UserID) (*iam.Recovery, error) {
	ex := r.executor(ctx)

	var (
		rawUserID string
		rawToken  string
		version   int
		createdAt time.Time
		expiresAt time.Time
	)

	err := ex.QueryRowContext(
		ctx,
		`SELECT
			user_id,
			token,
			version,
			created_at,
			expires_at
		FROM recoveries
		WHERE user_id = $1`,
		uID.String(),
	).Scan(
		&rawUserID,
		&rawToken,
		&version,
		&createdAt,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", iam.ErrRecoveryNotFound, err)
		}

		return nil, fmt.Errorf("iam.RecoveryRepository.GetByUser: %w", err)
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
		expiresAt,
	), nil
}

func (r *RecoveryRepository) Save(ctx context.Context, recovery *iam.Recovery) error {
	ex := r.executor(ctx)

	result, err := ex.ExecContext(
		ctx,
		`UPDATE recoveries
		SET
			token = $1,
			version = $2,
			created_at = $3,
			expires_at = $4
		WHERE
			user_id = $5
			AND version = $6`,
		recovery.Token().String(),
		recovery.Version()+1,
		recovery.CreatedAt(),
		recovery.ExpiresAt(),
		recovery.UserID().String(),
		recovery.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.RecoveryRepository.Save: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.RecoveryRepository.Save: rows affected: %w", err)
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
		WHERE user_id = $1
		  AND version = $2`,
		recovery.UserID().String(),
		recovery.Version(),
	)
	if err != nil {
		return fmt.Errorf("iam.RecoveryRepository.Remove: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("iam.RecoveryRepository.Remove: rows affected: %w", err)
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
		WHERE user_id = $1`,
		uID.String(),
	)
	if err != nil {
		return fmt.Errorf("iam.RecoveryRepository.RemoveByUser: %w", err)
	}

	if _, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("iam.RecoveryRepository.RemoveByUser: rows affected: %w", err)
	}

	return nil
}
