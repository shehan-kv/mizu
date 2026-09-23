package authz

import (
	"context"
	"errors"
	"testing"
	"time"

	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

type fakeUserRepository struct {
	getByIDFn func(context.Context, iam.UserID) (*iam.User, error)

	getByIDCalls int
}

var _ iam.UserRepository = (*fakeUserRepository)(nil)

func (f *fakeUserRepository) Add(
	ctx context.Context,
	user *iam.User,
) error {
	panic("unexpected call to Add")
}

func (f *fakeUserRepository) Exists(
	ctx context.Context,
	id iam.UserID,
) (bool, error) {
	panic("unexpected call to Exists")
}

func (f *fakeUserRepository) ExistsAll(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	panic("unexpected call to ExistsAll")
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id iam.UserID,
) (*iam.User, error) {
	f.getByIDCalls++

	return f.getByIDFn(ctx, id)
}

func (f *fakeUserRepository) GetByEmail(
	ctx context.Context,
	e iam.Email,
) (*iam.User, error) {
	panic("unexpected call to GetByEmail")
}

func (f *fakeUserRepository) List(
	ctx context.Context,
	filter iam.UserFilter,
	page common.Page,
) ([]*iam.User, error) {
	panic("unexpected call to List")
}

func (f *fakeUserRepository) ListByIDs(
	ctx context.Context,
	ids []iam.UserID,
	filter iam.UserFilter,
) ([]*iam.User, error) {
	panic("unexpected call to ListByIDs")
}

func (f *fakeUserRepository) IsAnyAdministrator(
	ctx context.Context,
	ids []iam.UserID,
) (bool, error) {
	panic("unexpected call to IsAnyAdministrator")
}

func (f *fakeUserRepository) HasAdministrator(
	ctx context.Context,
) (bool, error) {
	panic("unexpected call to HasAdministrator")
}

func (f *fakeUserRepository) Count(
	ctx context.Context,
	filter iam.UserFilter,
) (int, error) {
	panic("unexpected call to Count")
}

func (f *fakeUserRepository) Save(
	ctx context.Context,
	user *iam.User,
) error {
	panic("unexpected call to Save")
}

func (f *fakeUserRepository) Remove(
	ctx context.Context,
	user *iam.User,
) error {
	panic("unexpected call to Remove")
}

func newTestAuthzUser(t *testing.T, role iam.Role) *iam.User {
	t.Helper()

	now := time.Now()

	id, err := iam.NewUserID("user-1")
	if err != nil {
		t.Fatal(err)
	}

	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatal(err)
	}

	email, err := iam.NewEmail("test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	return iam.NewSystemUser(
		id,
		name,
		email,
		nil,
		role,
		true,
		now,
	)
}

func TestRequireAdministrator(t *testing.T) {
	t.Run("administrator is allowed", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleAdministrator), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministrator(
			context.Background(),
			actorID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("staff is forbidden", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleStaff), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministrator(
			context.Background(),
			actorID,
		)

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("client is forbidden", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleClient), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministrator(
			context.Background(),
			actorID,
		)

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repoErr := errors.New("repository failure")

		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, repoErr
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministrator(
			context.Background(),
			actorID,
		)

		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestRequireAdministratorOrSelf(t *testing.T) {
	t.Run("self is allowed without repository lookup", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, errors.New("repository should not be called")
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if repo.getByIDCalls != 0 {
			t.Fatalf(
				"expected repository not to be called, got %d calls",
				repo.getByIDCalls,
			)
		}
	})

	t.Run("administrator can act on another user", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleAdministrator), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("admin-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-administrator cannot act on another user", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleStaff), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("staff-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repoErr := errors.New("repository failure")

		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, repoErr
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("admin-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestRequireAdministratorOrStaff(t *testing.T) {
	tests := []struct {
		name        string
		role        iam.Role
		expectedErr error
	}{
		{
			name: "administrator is allowed",
			role: iam.RoleAdministrator,
		},
		{
			name: "staff is allowed",
			role: iam.RoleStaff,
		},
		{
			name:        "client is forbidden",
			role:        iam.RoleClient,
			expectedErr: ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				getByIDFn: func(
					context.Context,
					iam.UserID,
				) (*iam.User, error) {
					return newTestAuthzUser(t, tt.role), nil
				},
			}

			service := NewService(repo)

			actorID, err := iam.NewUserID("user-1")
			if err != nil {
				t.Fatal(err)
			}

			err = service.RequireAdministratorOrStaff(
				context.Background(),
				actorID,
			)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.expectedErr,
						err,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	t.Run("repository error is propagated", func(t *testing.T) {
		repoErr := errors.New("repository failure")

		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, repoErr
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorOrStaff(
			context.Background(),
			actorID,
		)

		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestRequireAdministratorStaffOrSelf(t *testing.T) {
	t.Run("self is allowed without repository lookup", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, errors.New("repository should not be called")
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorStaffOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if repo.getByIDCalls != 0 {
			t.Fatalf(
				"expected repository not to be called, got %d calls",
				repo.getByIDCalls,
			)
		}
	})

	t.Run("administrator can act on another user", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleAdministrator), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("admin-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorStaffOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("staff can act on another user", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleStaff), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("staff-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorStaffOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("client cannot act on another user", func(t *testing.T) {
		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return newTestAuthzUser(t, iam.RoleClient), nil
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("client-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorStaffOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repoErr := errors.New("repository failure")

		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, repoErr
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("admin-1")
		if err != nil {
			t.Fatal(err)
		}

		targetID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireAdministratorStaffOrSelf(
			context.Background(),
			actorID,
			targetID,
		)

		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}

func TestRequireClient(t *testing.T) {
	tests := []struct {
		name        string
		role        iam.Role
		expectedErr error
	}{
		{
			name: "client is allowed",
			role: iam.RoleClient,
		},
		{
			name:        "administrator is forbidden",
			role:        iam.RoleAdministrator,
			expectedErr: ErrForbidden,
		},
		{
			name:        "staff is forbidden",
			role:        iam.RoleStaff,
			expectedErr: ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				getByIDFn: func(
					context.Context,
					iam.UserID,
				) (*iam.User, error) {
					return newTestAuthzUser(t, tt.role), nil
				},
			}

			service := NewService(repo)

			actorID, err := iam.NewUserID("user-1")
			if err != nil {
				t.Fatal(err)
			}

			err = service.RequireClient(
				context.Background(),
				actorID,
			)

			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.expectedErr,
						err,
					)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	t.Run("repository error is propagated", func(t *testing.T) {
		repoErr := errors.New("repository failure")

		repo := &fakeUserRepository{
			getByIDFn: func(
				context.Context,
				iam.UserID,
			) (*iam.User, error) {
				return nil, repoErr
			},
		}

		service := NewService(repo)

		actorID, err := iam.NewUserID("user-1")
		if err != nil {
			t.Fatal(err)
		}

		err = service.RequireClient(
			context.Background(),
			actorID,
		)

		if !errors.Is(err, repoErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}
