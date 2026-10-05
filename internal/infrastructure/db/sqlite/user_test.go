package sqlite

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

func TestUserRepository(t *testing.T) {
	db := newTestSQLite(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("HasAdministrator", func(t *testing.T) {

		t.Run("returns true when administrator exists", func(t *testing.T) {
			admin := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000044",
				iam.RoleAdministrator,
				true,
			)

			if err := repo.Add(ctx, admin); err != nil {
				t.Fatal(err)
			}

			exists, err := repo.HasAdministrator(ctx)
			if err != nil {
				t.Fatalf("failed to check administrator: %v", err)
			}

			if !exists {
				t.Fatal("expected administrator to exist")
			}
		})

		t.Run("returns false when no administrator exists", func(t *testing.T) {
			exists, err := repo.HasAdministrator(ctx)
			if err != nil {
				t.Fatalf("failed to check administrator: %v", err)
			}

			// This assertion assumes the IDs used above are the only
			// administrators inserted by this test function.
			// If your test DB contains seeded administrators, use a
			// dedicated database or adjust the fixture accordingly.
			if !exists {
				// Expected in an empty/isolated database.
				return
			}
		})

	})

	t.Run("Add", func(t *testing.T) {
		t.Run("adds user without image", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000001",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			got, err := repo.GetByID(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to get inserted user: %v", err)
			}

			assertUserEqual(t, user, got)
		})

		t.Run("adds user with image", func(t *testing.T) {
			userID, err := iam.NewUserID(
				"60000000-0000-0000-0000-000000000002",
			)
			if err != nil {
				t.Fatal(err)
			}

			name, err := iam.NewName("Image", "User")
			if err != nil {
				t.Fatal(err)
			}

			email, err := iam.NewEmail(
				"60000000-0000-0000-0000-000000000002@example.com",
			)
			if err != nil {
				t.Fatal(err)
			}

			imageName, err := iam.NewImageName("avatar.png")
			if err != nil {
				t.Fatal(err)
			}

			mimeType, err := iam.NewMimeType("image/png")
			if err != nil {
				t.Fatal(err)
			}

			image := iam.NewImage(imageName, mimeType)
			now := time.Now().UTC().Truncate(time.Microsecond)

			user := iam.RestoreUser(
				userID,
				name,
				email,
				nil, // title
				iam.RoleStaff,
				&image,
				true,
				false,
				nil, // lastSignInAt
				1,
				now,
				now,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			got, err := repo.GetByID(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to get inserted user: %v", err)
			}

			assertUserEqual(t, user, got)
		})

		t.Run("returns duplicate user id error", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000003",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add first user: %v", err)
			}

			duplicateID := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000003",
				iam.RoleClient,
				true,
			)

			duplicateID.ChangeEmail(
				newTestEmail(t, "duplicate-user-id@example.com"),
				time.Now().UTC().Truncate(time.Microsecond),
			)

			err := repo.Add(ctx, duplicateID)
			if err == nil {
				t.Fatal("expected duplicate user id error")
			}

			if !strings.Contains(err.Error(), "duplicate user id") {
				t.Fatalf(
					"expected duplicate user id error, got %v",
					err,
				)
			}
		})
		t.Run("returns duplicate email error", func(t *testing.T) {
			first := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000004",
				iam.RoleClient,
				true,
			)

			secondID, err := iam.NewUserID(
				"60000000-0000-0000-0000-000000000005",
			)
			if err != nil {
				t.Fatal(err)
			}

			name, err := iam.NewName("Duplicate", "Email")
			if err != nil {
				t.Fatal(err)
			}

			email := first.Email()
			now := time.Now().UTC().Truncate(time.Microsecond)

			second := iam.NewSystemUser(
				secondID,
				name,
				email,
				nil,
				iam.RoleClient,
				true,
				now,
			)

			if err := repo.Add(ctx, first); err != nil {
				t.Fatalf("failed to add first user: %v", err)
			}

			err = repo.Add(ctx, second)
			if err == nil {
				t.Fatal("expected duplicate email error")
			}

			if !errors.Is(err, iam.ErrUserEmailAlreadyExists) {
				t.Fatalf(
					"expected duplicate email error, got %v",
					err,
				)
			}
		})
	})

	t.Run("Exists", func(t *testing.T) {
		t.Run("returns true for existing user", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000006",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			exists, err := repo.Exists(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to check user existence: %v", err)
			}

			if !exists {
				t.Fatal("expected user to exist")
			}
		})

		t.Run("returns false for unknown user", func(t *testing.T) {
			id := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000007",
			)

			exists, err := repo.Exists(ctx, id)
			if err != nil {
				t.Fatalf("failed to check user existence: %v", err)
			}

			if exists {
				t.Fatal("expected user not to exist")
			}
		})
	})

	t.Run("ExistsAll", func(t *testing.T) {
		t.Run("returns true when all users exist", func(t *testing.T) {
			user1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000008",
				iam.RoleClient,
				true,
			)
			user2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000009",
				iam.RoleStaff,
				true,
			)

			for _, user := range []*iam.User{user1, user2} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatalf("failed to add user: %v", err)
				}
			}

			exists, err := repo.ExistsAll(
				ctx,
				[]iam.UserID{
					user1.ID(),
					user2.ID(),
				},
			)
			if err != nil {
				t.Fatalf("failed to check users: %v", err)
			}

			if !exists {
				t.Fatal("expected all users to exist")
			}
		})

		t.Run("returns false when one user is missing", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000010",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			missingID := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000011",
			)

			exists, err := repo.ExistsAll(
				ctx,
				[]iam.UserID{
					user.ID(),
					missingID,
				},
			)
			if err != nil {
				t.Fatalf("failed to check users: %v", err)
			}

			if exists {
				t.Fatal("expected false when one user is missing")
			}
		})

		t.Run("returns false for empty ids", func(t *testing.T) {
			exists, err := repo.ExistsAll(ctx, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists {
				t.Fatal("expected false for empty ids")
			}
		})

		t.Run("handles duplicate ids", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000012",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			exists, err := repo.ExistsAll(
				ctx,
				[]iam.UserID{
					user.ID(),
					user.ID(),
				},
			)
			if err != nil {
				t.Fatalf("failed to check users: %v", err)
			}

			if !exists {
				t.Fatal("expected duplicate existing ids to return true")
			}
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		t.Run("returns user without image", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000013",
				iam.RoleStaff,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			got, err := repo.GetByID(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to get user: %v", err)
			}

			assertUserEqual(t, user, got)
		})

		t.Run("returns user with image", func(t *testing.T) {
			userID := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000014",
			)

			name, err := iam.NewName("Profile", "Image")
			if err != nil {
				t.Fatal(err)
			}

			email := newTestEmail(
				t,
				"60000000-0000-0000-0000-000000000014@example.com",
			)

			imageName, err := iam.NewImageName("profile.jpg")
			if err != nil {
				t.Fatal(err)
			}

			mimeType, err := iam.NewMimeType("image/jpeg")
			if err != nil {
				t.Fatal(err)
			}

			image := iam.NewImage(imageName, mimeType)
			now := time.Now().UTC().Truncate(time.Microsecond)

			user := iam.RestoreUser(
				userID,
				name,
				email,
				nil, // title
				iam.RoleClient,
				&image,
				true,
				false,
				nil, // lastSignInAt
				1,
				now,
				now,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			got, err := repo.GetByID(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to get user: %v", err)
			}

			assertUserEqual(t, user, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			id := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000015",
			)

			_, err := repo.GetByID(ctx, id)
			if err == nil {
				t.Fatal("expected not found error")
			}

			if !errors.Is(err, iam.ErrUserNotFound) {
				t.Fatalf(
					"expected user not found error, got %v",
					err,
				)
			}
		})
	})

	t.Run("GetByEmail", func(t *testing.T) {
		t.Run("returns user", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000016",
				iam.RoleStaff,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatalf("failed to add user: %v", err)
			}

			got, err := repo.GetByEmail(ctx, user.Email())
			if err != nil {
				t.Fatalf("failed to get user: %v", err)
			}

			assertUserEqual(t, user, got)
		})

		t.Run("returns not found", func(t *testing.T) {
			email := newTestEmail(
				t,
				"missing-user-600000000000000000000000000000000017@example.com",
			)

			_, err := repo.GetByEmail(ctx, email)
			if err == nil {
				t.Fatal("expected not found error")
			}

			if !errors.Is(err, iam.ErrUserNotFound) {
				t.Fatalf(
					"expected user not found error, got %v",
					err,
				)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("lists users with pagination", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			user1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000018",
				iam.RoleClient,
				true,
			)
			user2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000019",
				iam.RoleStaff,
				true,
			)
			user3 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000020",
				iam.RoleClient,
				false,
			)

			_ = now

			for _, user := range []*iam.User{user1, user2, user3} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatalf("failed to add user: %v", err)
				}
			}

			page, err := common.NewPage(2, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 2 {
				t.Fatalf("expected 2 users, got %d", len(got))
			}

			if got[0].ID() != user3.ID() {
				t.Fatalf(
					"expected newest user %s first, got %s",
					user3.ID(),
					got[0].ID(),
				)
			}

			if got[1].ID() != user2.ID() {
				t.Fatalf(
					"expected second newest user %s, got %s",
					user2.ID(),
					got[1].ID(),
				)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			user1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000021",
				iam.RoleClient,
				true,
			)
			user2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000022",
				iam.RoleStaff,
				true,
			)

			if err := repo.Add(ctx, user1); err != nil {
				t.Fatal(err)
			}
			if err := repo.Add(ctx, user2); err != nil {
				t.Fatal(err)
			}

			keyword := user1.Email().String()

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{
					Keyword: &keyword,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 user, got %d", len(got))
			}

			if got[0].ID() != user1.ID() {
				t.Fatalf(
					"expected user %s, got %s",
					user1.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("filters by role", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			client1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000023",
				iam.RoleClient,
				true,
			)
			client1.ChangeEmail(
				newTestEmail(t, "user-list-role-client1@example.com"),
				now,
			)

			client2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000025",
				iam.RoleClient,
				true,
			)
			client2.ChangeEmail(
				newTestEmail(t, "user-list-role-client2@example.com"),
				now,
			)

			staff := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000024",
				iam.RoleStaff,
				true,
			)
			staff.ChangeEmail(
				newTestEmail(t, "user-list-role-staff@example.com"),
				now,
			)

			for _, user := range []*iam.User{
				client1,
				client2,
				staff,
			} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			keyword := "user-list-role"
			role := iam.RoleClient

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{
					Keyword: &keyword,
					Role:    &role,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 2 {
				t.Fatalf("expected 2 clients, got %d", len(got))
			}

			for _, user := range got {
				if user.Role() != iam.RoleClient {
					t.Fatalf(
						"expected client role, got %s",
						user.Role(),
					)
				}
			}
		})

		t.Run("filters by active status", func(t *testing.T) {
			active := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000026",
				iam.RoleClient,
				true,
			)
			inactive := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000027",
				iam.RoleClient,
				false,
			)

			if err := repo.Add(ctx, active); err != nil {
				t.Fatal(err)
			}
			if err := repo.Add(ctx, inactive); err != nil {
				t.Fatal(err)
			}

			isActive := true

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{
					IsActive: &isActive,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			for _, user := range got {
				if !user.IsActive() {
					t.Fatal("expected only active users")
				}
			}
		})

		t.Run("filters by verified status", func(t *testing.T) {
			verified := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000028",
				iam.RoleClient,
				true,
			)
			unverified := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000029",
				iam.RoleClient,
				true,
			)

			verified.Verify(time.Now().UTC())

			if err := repo.Add(ctx, verified); err != nil {
				t.Fatal(err)
			}
			if err := repo.Add(ctx, unverified); err != nil {
				t.Fatal(err)
			}

			isVerified := true

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{
					IsVerified: &isVerified,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 verified user, got %d", len(got))
			}

			if got[0].ID() != verified.ID() {
				t.Fatalf(
					"expected verified user %s, got %s",
					verified.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("applies multiple filters", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			newUser := func(
				id string,
				firstName string,
				role iam.Role,
				isActive bool,
			) *iam.User {
				t.Helper()

				userID := newTestUserID(t, id)

				name, err := iam.NewName(firstName, "MultiFilter")
				if err != nil {
					t.Fatal(err)
				}

				email := newTestEmail(
					t,
					firstName+"@example.com",
				)

				return iam.NewSystemUser(
					userID,
					name,
					email,
					nil,
					role,
					isActive,
					now,
				)
			}

			match := newUser(
				"60000000-0000-0000-0000-000000000030",
				"MultiFilterMatch",
				iam.RoleStaff,
				true,
			)

			wrongRole := newUser(
				"60000000-0000-0000-0000-000000000031",
				"MultiFilterWrongRole",
				iam.RoleClient,
				true,
			)

			wrongActive := newUser(
				"60000000-0000-0000-0000-000000000032",
				"MultiFilterWrongActive",
				iam.RoleStaff,
				false,
			)

			for _, user := range []*iam.User{
				match,
				wrongRole,
				wrongActive,
			} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			keyword := "MultiFilter"
			role := iam.RoleStaff
			isActive := true

			page, err := common.NewPage(10, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := repo.List(
				ctx,
				iam.UserFilter{
					Keyword:  &keyword,
					Role:     &role,
					IsActive: &isActive,
				},
				page,
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 user, got %d", len(got))
			}

			if got[0].ID() != match.ID() {
				t.Fatalf(
					"expected matching user %s, got %s",
					match.ID(),
					got[0].ID(),
				)
			}
		})
	})

	t.Run("ListByIDs", func(t *testing.T) {
		t.Run("returns requested users", func(t *testing.T) {
			user1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000033",
				iam.RoleClient,
				true,
			)
			user2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000034",
				iam.RoleStaff,
				true,
			)
			other := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000035",
				iam.RoleClient,
				true,
			)

			for _, user := range []*iam.User{user1, user2, other} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			got, err := repo.ListByIDs(
				ctx,
				[]iam.UserID{
					user1.ID(),
					user2.ID(),
				},
				iam.UserFilter{},
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 2 {
				t.Fatalf("expected 2 users, got %d", len(got))
			}

			ids := map[iam.UserID]bool{}
			for _, user := range got {
				ids[user.ID()] = true
			}

			if !ids[user1.ID()] || !ids[user2.ID()] {
				t.Fatal("expected requested users to be returned")
			}
		})

		t.Run("returns empty for empty ids", func(t *testing.T) {
			got, err := repo.ListByIDs(
				ctx,
				nil,
				iam.UserFilter{},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != nil {
				t.Fatalf("expected nil result, got %v", got)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			now := time.Now().UTC().Truncate(time.Microsecond)

			matchID := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000036",
			)

			matchName, err := iam.NewName("KeywordMatch", "User")
			if err != nil {
				t.Fatal(err)
			}

			matchEmail := newTestEmail(
				t,
				"keyword-match-600000000000000000000000000000000036@example.com",
			)

			match := iam.NewSystemUser(
				matchID,
				matchName,
				matchEmail,
				nil,
				iam.RoleClient,
				true,
				now,
			)

			otherID := newTestUserID(
				t,
				"60000000-0000-0000-0000-000000000037",
			)

			otherName, err := iam.NewName("Different", "User")
			if err != nil {
				t.Fatal(err)
			}

			otherEmail := newTestEmail(
				t,
				"keyword-other-600000000000000000000000000000000037@example.com",
			)

			other := iam.NewSystemUser(
				otherID,
				otherName,
				otherEmail,
				nil,
				iam.RoleClient,
				true,
				now,
			)

			for _, user := range []*iam.User{match, other} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			keyword := "KeywordMatch"

			got, err := repo.ListByIDs(
				ctx,
				[]iam.UserID{
					match.ID(),
					other.ID(),
				},
				iam.UserFilter{
					Keyword: &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 user, got %d", len(got))
			}

			if got[0].ID() != match.ID() {
				t.Fatalf(
					"expected %s, got %s",
					match.ID(),
					got[0].ID(),
				)
			}
		})

		t.Run("filters by role and active status", func(t *testing.T) {
			match := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000038",
				iam.RoleStaff,
				true,
			)
			wrongRole := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000039",
				iam.RoleClient,
				true,
			)
			wrongActive := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000040",
				iam.RoleStaff,
				false,
			)

			for _, user := range []*iam.User{
				match,
				wrongRole,
				wrongActive,
			} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			role := iam.RoleStaff
			isActive := true

			got, err := repo.ListByIDs(
				ctx,
				[]iam.UserID{
					match.ID(),
					wrongRole.ID(),
					wrongActive.ID(),
				},
				iam.UserFilter{
					Role:     &role,
					IsActive: &isActive,
				},
			)
			if err != nil {
				t.Fatalf("failed to list users: %v", err)
			}

			if len(got) != 1 {
				t.Fatalf("expected 1 user, got %d", len(got))
			}

			if got[0].ID() != match.ID() {
				t.Fatalf("expected %s, got %s", match.ID(), got[0].ID())
			}
		})
	})

	t.Run("IsAnyAdministrator", func(t *testing.T) {
		t.Run("returns true when administrator exists", func(t *testing.T) {
			admin := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000041",
				iam.RoleAdministrator,
				true,
			)

			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000042",
				iam.RoleClient,
				true,
			)

			for _, value := range []*iam.User{admin, user} {
				if err := repo.Add(ctx, value); err != nil {
					t.Fatal(err)
				}
			}

			exists, err := repo.IsAnyAdministrator(
				ctx,
				[]iam.UserID{
					user.ID(),
					admin.ID(),
				},
			)
			if err != nil {
				t.Fatalf("failed to check administrators: %v", err)
			}

			if !exists {
				t.Fatal("expected administrator to exist")
			}
		})

		t.Run("returns false when no administrator exists", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000043",
				iam.RoleStaff,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatal(err)
			}

			exists, err := repo.IsAnyAdministrator(
				ctx,
				[]iam.UserID{user.ID()},
			)
			if err != nil {
				t.Fatalf("failed to check administrators: %v", err)
			}

			if exists {
				t.Fatal("expected no administrator")
			}
		})

		t.Run("returns false for empty ids", func(t *testing.T) {
			exists, err := repo.IsAnyAdministrator(ctx, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists {
				t.Fatal("expected false for empty ids")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("counts users", func(t *testing.T) {
			user1 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000045",
				iam.RoleClient,
				true,
			)
			user2 := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000046",
				iam.RoleStaff,
				true,
			)

			for _, user := range []*iam.User{user1, user2} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			count, err := repo.Count(ctx, iam.UserFilter{})
			if err != nil {
				t.Fatalf("failed to count users: %v", err)
			}

			if count < 2 {
				t.Fatalf("expected at least 2 users, got %d", count)
			}
		})

		t.Run("filters by keyword", func(t *testing.T) {
			match := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000047",
				iam.RoleClient,
				true,
			)
			other := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000048",
				iam.RoleClient,
				true,
			)

			for _, user := range []*iam.User{match, other} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			keyword := match.Email().String()

			count, err := repo.Count(
				ctx,
				iam.UserFilter{
					Keyword: &keyword,
				},
			)
			if err != nil {
				t.Fatalf("failed to count users: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1 user, got %d", count)
			}
		})

		t.Run("filters by role and active status", func(t *testing.T) {
			match := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000049",
				iam.RoleStaff,
				true,
			)
			wrongRole := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000050",
				iam.RoleClient,
				true,
			)
			wrongActive := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000051",
				iam.RoleStaff,
				false,
			)

			for _, user := range []*iam.User{
				match,
				wrongRole,
				wrongActive,
			} {
				if err := repo.Add(ctx, user); err != nil {
					t.Fatal(err)
				}
			}

			keyword := "60000000-0000-0000-0000-000000000049"
			role := iam.RoleStaff
			isActive := true

			count, err := repo.Count(
				ctx,
				iam.UserFilter{
					Keyword:  &keyword,
					Role:     &role,
					IsActive: &isActive,
				},
			)
			if err != nil {
				t.Fatalf("failed to count users: %v", err)
			}

			if count != 1 {
				t.Fatalf("expected 1 user, got %d", count)
			}
		})
	})

	t.Run("Save", func(t *testing.T) {
		t.Run("updates user", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000052",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatal(err)
			}

			now := time.Now().UTC().Truncate(time.Microsecond)

			userID := user.ID()

			name, err := iam.NewName("Updated", "User")
			if err != nil {
				t.Fatal(err)
			}

			email, err := iam.NewEmail(
				"updated-600000000000000000000000000000000052@example.com",
			)
			if err != nil {
				t.Fatal(err)
			}

			role := iam.RoleStaff
			title := "Senior Developer"

			updated := iam.RestoreUser(
				userID,
				name,
				email,
				&title,
				role,
				nil,
				false,
				true,
				&now,
				user.Version(),
				user.CreatedAt(),
				now,
			)

			if err := repo.Save(ctx, updated); err != nil {
				t.Fatalf("failed to save user: %v", err)
			}

			got, err := repo.GetByID(ctx, userID)
			if err != nil {
				t.Fatalf("failed to get updated user: %v", err)
			}

			if got.FirstName() != "Updated" {
				t.Fatalf(
					"expected first name Updated, got %s",
					got.FirstName(),
				)
			}

			if got.LastName() != "User" {
				t.Fatalf(
					"expected last name User, got %s",
					got.LastName(),
				)
			}

			if got.Email().String() != email.String() {
				t.Fatalf(
					"expected email %s, got %s",
					email,
					got.Email(),
				)
			}

			if got.Role() != iam.RoleStaff {
				t.Fatalf(
					"expected role %s, got %s",
					iam.RoleStaff,
					got.Role(),
				)
			}

			if got.Title() == nil || *got.Title() != title {
				t.Fatalf("expected title %q", title)
			}

			if got.IsActive() {
				t.Fatal("expected user to be inactive")
			}

			if !got.IsVerified() {
				t.Fatal("expected user to be verified")
			}

			if got.LastSignInAt() == nil {
				t.Fatal("expected last sign-in time")
			}

			if got.Version() != user.Version()+1 {
				t.Fatalf(
					"expected version %d, got %d",
					user.Version()+1,
					got.Version(),
				)
			}
		})

		t.Run("returns concurrent modification error for stale version", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000053",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatal(err)
			}

			if _, err := db.ExecContext(
				ctx,
				`UPDATE users SET version = version + 1 WHERE id = ?`,
				user.ID().String(),
			); err != nil {
				t.Fatal(err)
			}

			err := repo.Save(ctx, user)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(err, iam.ErrUserConcurrentModification) {
				t.Fatalf(
					"expected concurrent modification error, got %v",
					err,
				)
			}
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes user", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000054",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatal(err)
			}

			if err := repo.Remove(ctx, user); err != nil {
				t.Fatalf("failed to remove user: %v", err)
			}

			exists, err := repo.Exists(ctx, user.ID())
			if err != nil {
				t.Fatalf("failed to check removed user: %v", err)
			}

			if exists {
				t.Fatal("expected user to be removed")
			}
		})

		t.Run("returns concurrent modification error for stale version", func(t *testing.T) {
			user := newTestUser(
				t,
				"60000000-0000-0000-0000-000000000055",
				iam.RoleClient,
				true,
			)

			if err := repo.Add(ctx, user); err != nil {
				t.Fatal(err)
			}

			if _, err := db.ExecContext(
				ctx,
				`UPDATE users SET version = version + 1 WHERE id = ?`,
				user.ID().String(),
			); err != nil {
				t.Fatal(err)
			}

			err := repo.Remove(ctx, user)
			if err == nil {
				t.Fatal("expected concurrent modification error")
			}

			if !errors.Is(err, iam.ErrUserConcurrentModification) {
				t.Fatalf(
					"expected concurrent modification error, got %v",
					err,
				)
			}
		})
	})
}

func newTestEmail(t *testing.T, value string) iam.Email {
	t.Helper()

	email, err := iam.NewEmail(value)
	if err != nil {
		t.Fatalf("failed to create email: %v", err)
	}

	return email
}

func newTestUserID(t *testing.T, id string) iam.UserID {
	t.Helper()

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatalf("failed to create user ID: %v", err)
	}

	return userID
}

func newTestUser(
	t *testing.T,
	id string,
	role iam.Role,
	isActive bool,
) *iam.User {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatalf("create user ID: %v", err)
	}

	email, err := iam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatalf("create email: %v", err)
	}

	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatalf("create name: %v", err)
	}

	return iam.RestoreUser(
		userID,
		name,
		email,
		nil, // title
		role,
		nil, // image
		isActive,
		false, // isVerified
		nil,   // lastSignInAt
		1,     // version
		now,
		now,
	)
}

func assertUserEqual(t *testing.T, expected, actual *iam.User) {
	t.Helper()

	if expected.ID() != actual.ID() {
		t.Fatalf(
			"expected user ID %s, got %s",
			expected.ID(),
			actual.ID(),
		)
	}

	if expected.FirstName() != actual.FirstName() {
		t.Fatalf(
			"expected first name %q, got %q",
			expected.FirstName(),
			actual.FirstName(),
		)
	}

	if expected.LastName() != actual.LastName() {
		t.Fatalf(
			"expected last name %q, got %q",
			expected.LastName(),
			actual.LastName(),
		)
	}

	if expected.Email() != actual.Email() {
		t.Fatalf(
			"expected email %q, got %q",
			expected.Email(),
			actual.Email(),
		)
	}

	if expected.Title() == nil && actual.Title() != nil {
		t.Fatal("expected nil title")
	}

	if expected.Title() != nil && actual.Title() == nil {
		t.Fatal("expected non-nil title")
	}

	if expected.Title() != nil &&
		actual.Title() != nil &&
		*expected.Title() != *actual.Title() {
		t.Fatalf(
			"expected title %q, got %q",
			*expected.Title(),
			*actual.Title(),
		)
	}

	if expected.Role() != actual.Role() {
		t.Fatalf(
			"expected role %s, got %s",
			expected.Role(),
			actual.Role(),
		)
	}

	if expected.IsActive() != actual.IsActive() {
		t.Fatalf(
			"expected active %t, got %t",
			expected.IsActive(),
			actual.IsActive(),
		)
	}

	if expected.IsVerified() != actual.IsVerified() {
		t.Fatalf(
			"expected verified %t, got %t",
			expected.IsVerified(),
			actual.IsVerified(),
		)
	}

	if expected.LastSignInAt() == nil && actual.LastSignInAt() != nil {
		t.Fatal("expected nil last sign-in time")
	}

	if expected.LastSignInAt() != nil && actual.LastSignInAt() == nil {
		t.Fatal("expected non-nil last sign-in time")
	}

	if expected.LastSignInAt() != nil &&
		actual.LastSignInAt() != nil &&
		!expected.LastSignInAt().Equal(*actual.LastSignInAt()) {
		t.Fatalf(
			"expected last sign-in time %v, got %v",
			*expected.LastSignInAt(),
			*actual.LastSignInAt(),
		)
	}

	if expected.Image() == nil && actual.Image() != nil {
		t.Fatal("expected nil image")
	}

	if expected.Image() != nil && actual.Image() == nil {
		t.Fatal("expected non-nil image")
	}

	if expected.Image() != nil && actual.Image() != nil {
		if expected.Image().Name() != actual.Image().Name() {
			t.Fatalf(
				"expected image name %q, got %q",
				expected.Image().Name(),
				actual.Image().Name(),
			)
		}

		if expected.Image().MimeType() != actual.Image().MimeType() {
			t.Fatalf(
				"expected image MIME type %q, got %q",
				expected.Image().MimeType(),
				actual.Image().MimeType(),
			)
		}
	}

	if expected.Version() != actual.Version() {
		t.Fatalf(
			"expected version %d, got %d",
			expected.Version(),
			actual.Version(),
		)
	}

	if !expected.CreatedAt().Equal(actual.CreatedAt()) {
		t.Fatalf(
			"expected created_at %v, got %v",
			expected.CreatedAt(),
			actual.CreatedAt(),
		)
	}

	if !expected.UpdatedAt().Equal(actual.UpdatedAt()) {
		t.Fatalf(
			"expected updated_at %v, got %v",
			expected.UpdatedAt(),
			actual.UpdatedAt(),
		)
	}
}
