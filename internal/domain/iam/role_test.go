package iam

import "testing"

func TestNewRole(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Role
		wantError error
	}{
		{
			name:  "administrator",
			input: "administrator",
			want:  RoleAdministrator,
		},
		{
			name:  "staff",
			input: "staff",
			want:  RoleStaff,
		},
		{
			name:  "client",
			input: "client",
			want:  RoleClient,
		},
		{
			name:      "invalid role",
			input:     "manager",
			wantError: ErrUserInvalidRole,
		},
		{
			name:      "empty role",
			input:     "",
			wantError: ErrUserInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := NewRole(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if role != tt.want {
				t.Errorf("expected role %q, got %q", tt.want, role)
			}

			if role.String() != tt.input {
				t.Errorf("expected String() to return %q, got %q", tt.input, role.String())
			}
		})
	}
}
