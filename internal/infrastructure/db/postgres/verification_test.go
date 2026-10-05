package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestVerificationRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewVerificationRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds and retrieves verification", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000001",
			)

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000002",
			)
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now().UTC().Truncate(time.Microsecond)

			verification := iam.NewVerification(
				verificationID,
				userID,
				now,
			)

			if err := repo.Add(ctx, verification); err != nil {
				t.Fatalf("failed to add verification: %v", err)
			}

			got, err := repo.Get(ctx, verificationID)
			if err != nil {
				t.Fatalf("failed to retrieve verification: %v", err)
			}

			assertVerificationEqual(t, verification, got)
		})

		t.Run("returns foreign key error for unknown user", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"00000000-0000-0000-0000-000000000099",
			)
			if err != nil {
				t.Fatal(err)
			}

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000003",
			)
			if err != nil {
				t.Fatal(err)
			}

			verification := iam.NewVerification(
				verificationID,
				userID,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			err = repo.Add(ctx, verification)
			if err == nil {
				t.Fatal("expected foreign key error")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns verification", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000004",
			)

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000005",
			)
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now().UTC().Truncate(time.Microsecond)

			verification := iam.NewVerification(
				verificationID,
				userID,
				now,
			)

			if err := repo.Add(ctx, verification); err != nil {
				t.Fatalf("failed to add verification: %v", err)
			}

			got, err := repo.Get(ctx, verificationID)
			if err != nil {
				t.Fatalf("failed to get verification: %v", err)
			}

			assertVerificationEqual(t, verification, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000006",
			)
			if err != nil {
				t.Fatal(err)
			}

			_, err = repo.Get(ctx, verificationID)
			if !errors.Is(err, iam.ErrVerificationNotFound) {
				t.Fatalf(
					"expected verification-not-found error, got %v",
					err,
				)
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes verification", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000007",
			)

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000008",
			)
			if err != nil {
				t.Fatal(err)
			}

			verification := iam.NewVerification(
				verificationID,
				userID,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, verification); err != nil {
				t.Fatalf("failed to add verification: %v", err)
			}

			if err := repo.Remove(ctx, verification); err != nil {
				t.Fatalf("failed to remove verification: %v", err)
			}

			_, err = repo.Get(ctx, verificationID)
			if !errors.Is(err, iam.ErrVerificationNotFound) {
				t.Fatalf(
					"expected verification-not-found error, got %v",
					err,
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000009",
			)

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000010",
			)
			if err != nil {
				t.Fatal(err)
			}

			verification := iam.NewVerification(
				verificationID,
				userID,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, verification); err != nil {
				t.Fatalf("failed to add verification: %v", err)
			}

			stale := iam.RestoreVerification(
				verification.ID(),
				verification.UserID(),
				verification.Version()+1,
				verification.CreatedAt(),
			)

			err = repo.Remove(ctx, stale)
			if !errors.Is(err, iam.ErrVerificationConcurrentModification) {
				t.Fatalf(
					"expected concurrent modification error, got %v",
					err,
				)
			}

			got, err := repo.Get(ctx, verificationID)
			if err != nil {
				t.Fatalf(
					"expected verification to still exist, got %v",
					err,
				)
			}

			assertVerificationEqual(t, verification, got)
		})
	})

	t.Run("RemoveByUserID", func(t *testing.T) {
		t.Run("removes user's verification", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000011",
			)

			verificationID, err := iam.NewVerificationID(
				"00000000-0000-0000-0000-000000000012",
			)
			if err != nil {
				t.Fatal(err)
			}

			verification := iam.NewVerification(
				verificationID,
				userID,
				time.Now().UTC().Truncate(time.Microsecond),
			)

			if err := repo.Add(ctx, verification); err != nil {
				t.Fatalf("failed to add verification: %v", err)
			}

			if err := repo.RemoveByUserID(ctx, userID); err != nil {
				t.Fatalf(
					"failed to remove verification by user ID: %v",
					err,
				)
			}

			_, err = repo.Get(ctx, verificationID)
			if !errors.Is(err, iam.ErrVerificationNotFound) {
				t.Fatalf(
					"expected verification-not-found error, got %v",
					err,
				)
			}
		})

		t.Run("returns nil when user has no verification", func(t *testing.T) {
			userID := newTestVerificationUser(
				t,
				db,
				"00000000-0000-0000-0000-000000000013",
			)

			if err := repo.RemoveByUserID(ctx, userID); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	})
}

func newTestVerificationUser(
	t *testing.T,
	db *sql.DB,
	id string,
) iam.UserID {
	t.Helper()

	_, err := db.ExecContext(
		context.Background(),
		`INSERT INTO users(
			id,
			first_name,
			last_name,
			email,
			role
		) VALUES ($1, $2, $3, $4, $5)`,
		id,
		"Test",
		"User",
		id+"@example.com",
		"client",
	)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func assertVerificationEqual(
	t *testing.T,
	want *iam.Verification,
	got *iam.Verification,
) {
	t.Helper()

	if got.ID() != want.ID() {
		t.Fatalf(
			"expected ID %q, got %q",
			want.ID(),
			got.ID(),
		)
	}

	if got.UserID() != want.UserID() {
		t.Fatalf(
			"expected user ID %q, got %q",
			want.UserID(),
			got.UserID(),
		)
	}

	if got.Version() != want.Version() {
		t.Fatalf(
			"expected version %d, got %d",
			want.Version(),
			got.Version(),
		)
	}

	if !got.CreatedAt().Equal(want.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			want.CreatedAt(),
			got.CreatedAt(),
		)
	}
}
