package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func TestRecoveryRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewRecoveryRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds recovery", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"11111111-1111-1111-1111-111111111111",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-1",
			)

			err := repo.Add(ctx, recovery)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("failed to get inserted recovery: %v", err)
			}

			assertRecoveryEqual(t, recovery, got)
		})

		t.Run("returns duplicate recovery error", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"22222222-2222-2222-2222-222222222222",
			)

			first := newTestRecovery(
				t,
				userID,
				"recovery-token-2",
			)

			if err := repo.Add(ctx, first); err != nil {
				t.Fatalf("failed to add first recovery: %v", err)
			}

			second := newTestRecovery(
				t,
				userID,
				"recovery-token-3",
			)

			err := repo.Add(ctx, second)
			if err == nil {
				t.Fatal("expected duplicate recovery error")
			}

			if !strings.Contains(err.Error(), "duplicate recovery") {
				t.Fatalf(
					"expected duplicate recovery error, got %v",
					err,
				)
			}
		})
		t.Run("returns foreign key violation", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"33333333-3333-3333-3333-333333333333",
			)
			if err != nil {
				t.Fatal(err)
			}

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-3",
			)

			err = repo.Add(ctx, recovery)
			if err == nil {
				t.Fatal("expected foreign key violation")
			}

			if !strings.Contains(
				err.Error(),
				"foreign key constraint violation",
			) {
				t.Fatalf(
					"expected foreign key violation, got %v",
					err,
				)
			}
		})
	})

	t.Run("GetByToken", func(t *testing.T) {
		t.Run("returns recovery", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"44444444-4444-4444-4444-444444444444",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-4",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			got, err := repo.GetByToken(ctx, recovery.Token())
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			assertRecoveryEqual(t, recovery, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			token, err := iam.NewRecoveryToken("missing-token")
			if err != nil {
				t.Fatal(err)
			}

			recovery, err := repo.GetByToken(ctx, token)
			if recovery != nil {
				t.Fatal("expected nil recovery")
			}

			if !errors.Is(err, iam.ErrRecoveryNotFound) {
				t.Fatalf(
					"expected ErrRecoveryNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("GetByUser", func(t *testing.T) {
		t.Run("returns recovery", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"55555555-5555-5555-5555-555555555555",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-5",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			assertRecoveryEqual(t, recovery, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"66666666-6666-6666-6666-666666666666",
			)
			if err != nil {
				t.Fatal(err)
			}

			recovery, err := repo.GetByUser(ctx, userID)
			if recovery != nil {
				t.Fatal("expected nil recovery")
			}

			if !errors.Is(err, iam.ErrRecoveryNotFound) {
				t.Fatalf(
					"expected ErrRecoveryNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates recovery and increments version", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"77777777-7777-7777-7777-777777777777",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"old-recovery-token",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			newToken, err := iam.NewRecoveryToken("new-recovery-token")
			if err != nil {
				t.Fatal(err)
			}

			updatedAt := time.Now().UTC().Truncate(time.Microsecond)

			// Recovery does not expose a mutation for the token itself
			// in the test fixture, so restore a new aggregate with the
			// same persisted version.
			updated := iam.RestoreRecovery(
				userID,
				newToken,
				recovery.Version(),
				updatedAt,
			)

			err = repo.Save(ctx, updated)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("failed to retrieve saved recovery: %v", err)
			}

			if got.Token() != newToken {
				t.Fatalf(
					"expected token %q, got %q",
					newToken,
					got.Token(),
				)
			}

			if got.Version() != recovery.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					recovery.Version()+1,
					got.Version(),
				)
			}

			if !got.CreatedAt().Equal(updatedAt) {
				t.Fatalf(
					"expected created_at %v, got %v",
					updatedAt,
					got.CreatedAt(),
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"88888888-8888-8888-8888-888888888888",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-8",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			// Version 2 does not match the persisted version 1.
			newToken, err := iam.NewRecoveryToken("new-recovery-token-8")
			if err != nil {
				t.Fatal(err)
			}

			stale := iam.RestoreRecovery(
				userID,
				newToken,
				recovery.Version()+1,
				recovery.CreatedAt(),
			)

			err = repo.Save(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				iam.ErrRecoveryConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrRecoveryConcurrentModification, got %v",
					err,
				)
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes recovery", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"99999999-9999-9999-9999-999999999999",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-9",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			err := repo.Remove(ctx, recovery)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			_, err = repo.GetByUser(ctx, userID)
			if !errors.Is(err, iam.ErrRecoveryNotFound) {
				t.Fatalf(
					"expected ErrRecoveryNotFound after removal, got %v",
					err,
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-a",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			stale := iam.RestoreRecovery(
				userID,
				recovery.Token(),
				recovery.Version()+1,
				recovery.CreatedAt(),
			)

			err := repo.Remove(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				iam.ErrRecoveryConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrRecoveryConcurrentModification, got %v",
					err,
				)
			}
		})
	})

	t.Run("RemoveByUser", func(t *testing.T) {
		t.Run("removes user's recovery", func(t *testing.T) {
			userID := newTestRecoveryUser(
				t,
				db,
				"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
			)

			recovery := newTestRecovery(
				t,
				userID,
				"recovery-token-b",
			)

			if err := repo.Add(ctx, recovery); err != nil {
				t.Fatalf("failed to add recovery: %v", err)
			}

			err := repo.RemoveByUser(ctx, userID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			_, err = repo.GetByUser(ctx, userID)
			if !errors.Is(err, iam.ErrRecoveryNotFound) {
				t.Fatalf(
					"expected ErrRecoveryNotFound after removal, got %v",
					err,
				)
			}
		})

		t.Run("returns nil when user has no recovery", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"cccccccc-cccc-cccc-cccc-cccccccccccc",
			)
			if err != nil {
				t.Fatal(err)
			}

			err = repo.RemoveByUser(ctx, userID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	})
}

func newTestRecoveryUser(
	t *testing.T,
	db *sql.DB,
	id string,
) iam.UserID {
	t.Helper()

	insertTestRecoveryUser(t, db, id)

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	return userID
}

func insertTestRecoveryUser(
	t *testing.T,
	db *sql.DB,
	id string,
) {
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
}

func newTestRecovery(
	t *testing.T,
	userID iam.UserID,
	token string,
) *iam.Recovery {
	t.Helper()

	recoveryToken, err := iam.NewRecoveryToken(token)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return iam.NewRecovery(
		userID,
		recoveryToken,
		now,
	)
}

func assertRecoveryEqual(
	t *testing.T,
	expected *iam.Recovery,
	actual *iam.Recovery,
) {
	t.Helper()

	if actual == nil {
		t.Fatal("expected recovery, got nil")
	}

	if actual.UserID() != expected.UserID() {
		t.Fatalf(
			"expected user ID %q, got %q",
			expected.UserID(),
			actual.UserID(),
		)
	}

	if actual.Token() != expected.Token() {
		t.Fatalf(
			"expected token %q, got %q",
			expected.Token(),
			actual.Token(),
		)
	}

	if actual.Version() != expected.Version() {
		t.Fatalf(
			"expected version %d, got %d",
			expected.Version(),
			actual.Version(),
		)
	}

	if !actual.CreatedAt().Equal(expected.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			expected.CreatedAt(),
			actual.CreatedAt(),
		)
	}
}
