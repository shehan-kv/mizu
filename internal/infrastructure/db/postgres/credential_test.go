package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"mizu/internal/domain/iam"
)

func TestCredentialRepository(t *testing.T) {
	db := newTestPostgres(t)
	repo := NewCredentialRepository(db)
	ctx := context.Background()

	t.Run("Add", func(t *testing.T) {
		t.Run("adds credential", func(t *testing.T) {
			userID := newTestCredentialUser(
				t,
				db,
				"20000000-0000-0000-0000-000000000001",
			)

			credential := iam.NewCredential(
				userID,
				"hashed-password",
			)

			if err := repo.Add(ctx, credential); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("failed to get inserted credential: %v", err)
			}

			if got == nil {
				t.Fatal("expected credential, got nil")
			}

			if got.UserID() != credential.UserID() {
				t.Fatalf(
					"expected user ID %q, got %q",
					credential.UserID(),
					got.UserID(),
				)
			}

			if got.Hash() != credential.Hash() {
				t.Fatalf(
					"expected hash %q, got %q",
					credential.Hash(),
					got.Hash(),
				)
			}

			if got.Version() != credential.Version() {
				t.Fatalf(
					"expected version %d, got %d",
					credential.Version(),
					got.Version(),
				)
			}
		})

		t.Run("returns duplicate credential error", func(t *testing.T) {
			userID := newTestCredentialUser(
				t,
				db,
				"20000000-0000-0000-0000-000000000011",
			)

			first := iam.NewCredential(
				userID,
				"hashed-password-1",
			)

			if err := repo.Add(ctx, first); err != nil {
				t.Fatalf("failed to add first credential: %v", err)
			}

			second := iam.NewCredential(
				userID,
				"hashed-password-2",
			)

			err := repo.Add(ctx, second)
			if err == nil {
				t.Fatal("expected duplicate credential error")
			}

			if !contains(err.Error(), "duplicate credential") {
				t.Fatalf(
					"expected duplicate credential error, got %v",
					err,
				)
			}
		})

		t.Run("returns user does not exist error", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"20000000-0000-0000-0000-000000000021",
			)
			if err != nil {
				t.Fatal(err)
			}

			credential := iam.NewCredential(
				userID,
				"hashed-password",
			)

			err = repo.Add(ctx, credential)
			if err == nil {
				t.Fatal("expected user does not exist error")
			}

			if !contains(err.Error(), "user does not exist") {
				t.Fatalf(
					"expected user does not exist error, got %v",
					err,
				)
			}
		})
	})

	t.Run("GetByUser", func(t *testing.T) {
		t.Run("returns credential", func(t *testing.T) {
			userID := newTestCredentialUser(
				t,
				db,
				"20000000-0000-0000-0000-000000000031",
			)

			credential := iam.NewCredential(
				userID,
				"hashed-password",
			)

			if err := repo.Add(ctx, credential); err != nil {
				t.Fatalf("failed to add credential: %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			if got == nil {
				t.Fatal("expected credential, got nil")
			}

			if got.UserID() != credential.UserID() {
				t.Fatalf(
					"expected user ID %q, got %q",
					credential.UserID(),
					got.UserID(),
				)
			}

			if got.Hash() != credential.Hash() {
				t.Fatalf(
					"expected hash %q, got %q",
					credential.Hash(),
					got.Hash(),
				)
			}

			if got.Version() != credential.Version() {
				t.Fatalf(
					"expected version %d, got %d",
					credential.Version(),
					got.Version(),
				)
			}
		})

		t.Run("returns not found", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"20000000-0000-0000-0000-000000000041",
			)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.GetByUser(ctx, userID)

			if got != nil {
				t.Fatal("expected nil credential")
			}

			if !errors.Is(err, iam.ErrCredentialNotFound) {
				t.Fatalf(
					"expected ErrCredentialNotFound, got %v",
					err,
				)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates credential hash", func(t *testing.T) {
			userID := newTestCredentialUser(
				t,
				db,
				"20000000-0000-0000-0000-000000000051",
			)

			credential := iam.NewCredential(
				userID,
				"old-hashed-password",
			)

			if err := repo.Add(ctx, credential); err != nil {
				t.Fatalf("failed to add credential: %v", err)
			}

			credential.UpdateHash("new-hashed-password")

			if err := repo.Save(ctx, credential); err != nil {
				t.Fatalf("expected nil, got %v", err)
			}

			got, err := repo.GetByUser(ctx, userID)
			if err != nil {
				t.Fatalf("failed to retrieve saved credential: %v", err)
			}

			if got.Hash() != "new-hashed-password" {
				t.Fatalf(
					"expected updated hash %q, got %q",
					"new-hashed-password",
					got.Hash(),
				)
			}

			if got.Version() != credential.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					credential.Version()+1,
					got.Version(),
				)
			}
		})

		t.Run("returns concurrent modification error", func(t *testing.T) {
			userID := newTestCredentialUser(
				t,
				db,
				"20000000-0000-0000-0000-000000000061",
			)

			credential := iam.NewCredential(
				userID,
				"hashed-password",
			)

			if err := repo.Add(ctx, credential); err != nil {
				t.Fatalf("failed to add credential: %v", err)
			}

			stale := iam.RestoreCredential(
				credential.UserID(),
				"stale-hashed-password",
				credential.Version()+1,
			)

			err := repo.Save(ctx, stale)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(
				err,
				iam.ErrCredentialConcurrentModification,
			) {
				t.Fatalf(
					"expected ErrCredentialConcurrentModification, got %v",
					err,
				)
			}
		})
	})
}

func newTestCredentialUser(
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
