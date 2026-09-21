package iam

import "testing"

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError error
	}{
		{
			name:  "valid email",
			input: "user@example.com",
			want:  "user@example.com",
		},
		{
			name:      "empty email",
			input:     "",
			wantError: ErrUserEmailCannotBeEmpty,
		},
		{
			name:      "invalid email",
			input:     "not-an-email",
			wantError: ErrUserInvalidEmailAddress,
		},
		{
			name:      "invalid email without domain",
			input:     "user@",
			wantError: ErrUserInvalidEmailAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if email.String() != tt.want {
				t.Errorf("expected email %q, got %q", tt.want, email.String())
			}
		})
	}
}

func TestEmailString(t *testing.T) {
	email, err := NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := email.String(); got != "user@example.com" {
		t.Errorf("expected %q, got %q", "user@example.com", got)
	}
}
