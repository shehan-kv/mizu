package iam

import "testing"

func TestNewUserID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      UserID
		wantError error
	}{
		{
			name:  "valid ID",
			input: "user-123",
			want:  UserID("user-123"),
		},
		{
			name:      "empty ID",
			input:     "",
			wantError: ErrUserIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewUserID(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if id != tt.want {
				t.Errorf("expected ID %q, got %q", tt.want, id)
			}
		})
	}
}

func TestUserIDString(t *testing.T) {
	id := UserID("user-123")

	if got := id.String(); got != "user-123" {
		t.Errorf("expected %q, got %q", "user-123", got)
	}
}

func TestSystemUserID(t *testing.T) {
	if SystemUserID.String() == "" {
		t.Error("expected SystemUserID to be non-empty")
	}
}
