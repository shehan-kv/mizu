package iam

import (
	"context"
	"testing"
	"time"
)

type fakePasswordHasher struct {
	verifyResult bool

	verifyCalled     bool
	verifiedHash     string
	verifiedPassword PlainPassword
}

func (f *fakePasswordHasher) Hash(plain PlainPassword) (string, error) {
	return "", nil
}

func (f *fakePasswordHasher) Verify(hashed string, password PlainPassword) bool {
	f.verifyCalled = true
	f.verifiedHash = hashed
	f.verifiedPassword = password

	return f.verifyResult
}

type fakeCredentialRepository struct {
	credential *Credential
	err        error

	getByUserCalled bool
	gotUserID       UserID
}

func (f *fakeCredentialRepository) Add(
	ctx context.Context,
	c *Credential,
) error {
	return nil
}

func (f *fakeCredentialRepository) GetByUser(
	ctx context.Context,
	userID UserID,
) (*Credential, error) {
	f.getByUserCalled = true
	f.gotUserID = userID

	return f.credential, f.err
}

func (f *fakeCredentialRepository) Save(
	ctx context.Context,
	c *Credential,
) error {
	return nil
}

func TestServiceAuthenticate(t *testing.T) {

	now := time.Now()

	userID := UserID("user-123")
	name, err := NewName("John", "Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	email, err := NewEmail("john@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	role := RoleStaff

	password, err := NewPlainPassword("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name          string
		active        bool
		verifyResult  bool
		repositoryErr error
		wantError     error
	}{
		{
			name:         "successful authentication",
			active:       true,
			verifyResult: true,
		},
		{
			name:         "invalid credentials",
			active:       true,
			verifyResult: false,
			wantError:    ErrUserInvalidCredentials,
		},
		{
			name:         "inactive user",
			active:       false,
			verifyResult: true,
			wantError:    ErrUserInactive,
		},
		{
			name:          "credential repository error",
			active:        true,
			verifyResult:  true,
			repositoryErr: context.DeadlineExceeded,
			wantError:     context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(
				userID,
				name,
				email,
				nil,
				role,
				tt.active,
				UserID("actor"),
				now,
			)

			hasher := &fakePasswordHasher{
				verifyResult: tt.verifyResult,
			}

			repository := &fakeCredentialRepository{
				credential: NewCredential(userID, "stored-hash"),
				err:        tt.repositoryErr,
			}

			service := NewService(hasher, repository)

			err := service.Authenticate(
				context.Background(),
				user,
				password,
			)

			if err != tt.wantError {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantError,
					err,
				)
			}

			if !repository.getByUserCalled {
				t.Error("expected credential repository to be called")
			}

			if repository.gotUserID != userID {
				t.Errorf(
					"expected repository user ID %q, got %q",
					userID,
					repository.gotUserID,
				)
			}

			if tt.repositoryErr != nil {
				if hasher.verifyCalled {
					t.Error("expected password verification not to occur when repository fails")
				}

				return
			}

			if !hasher.verifyCalled {
				t.Error("expected password verification to be called")
			}

			if hasher.verifiedHash != "stored-hash" {
				t.Errorf(
					"expected hash %q, got %q",
					"stored-hash",
					hasher.verifiedHash,
				)
			}

			if !hasher.verifiedPassword.Equals(password) {
				t.Error("expected the supplied password to be passed to the hasher")
			}
		})
	}
}
