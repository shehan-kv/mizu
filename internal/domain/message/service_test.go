package message

import (
	"testing"
	"time"

	"mizu/internal/domain/iam"
)

func newTestMessageServiceUser(
	t *testing.T,
	id string,
	role iam.Role,
) *iam.User {
	t.Helper()

	userID, err := iam.NewUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	name, err := iam.NewName("Test", "User")
	if err != nil {
		t.Fatal(err)
	}

	email, err := iam.NewEmail(id + "@example.com")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	return iam.NewSystemUser(
		userID,
		name,
		email,
		nil,
		role,
		true,
		now,
	)
}

func TestServiceValidateStandaloneChannelMembers(t *testing.T) {
	service := NewService()

	tests := []struct {
		name    string
		members []*iam.User
		wantErr error
	}{
		{
			name: "administrator is allowed",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleAdministrator),
			},
		},
		{
			name: "staff is allowed",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleStaff),
			},
		},
		{
			name: "client is rejected",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleClient),
			},
			wantErr: ErrChannelMustHaveStaffMember,
		},
		{
			name: "administrator among clients is allowed",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleClient),
				newTestMessageServiceUser(t, "user-2", iam.RoleAdministrator),
			},
		},
		{
			name: "staff among clients is allowed",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleClient),
				newTestMessageServiceUser(t, "user-2", iam.RoleStaff),
			},
		},
		{
			name: "only clients are rejected",
			members: []*iam.User{
				newTestMessageServiceUser(t, "user-1", iam.RoleClient),
				newTestMessageServiceUser(t, "user-2", iam.RoleClient),
			},
			wantErr: ErrChannelMustHaveStaffMember,
		},
		{
			name:    "empty members are rejected",
			members: nil,
			wantErr: ErrChannelMustHaveStaffMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateStandaloneChannelMembers(tt.members)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
