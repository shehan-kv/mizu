package iam

import "testing"

func TestNewPlainPassword(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{
			name:  "minimum length password",
			input: "12345678",
		},
		{
			name:  "normal password",
			input: "secure-password",
		},
		{
			name:  "maximum bcrypt password length",
			input: "123456789012345678901234567890123456789012345678901234567890123456789012",
		},
		{
			name:      "password too short",
			input:     "1234567",
			wantError: ErrUserPasswordTooShort,
		},
		{
			name:      "password too long",
			input:     "1234567890123456789012345678901234567890123456789012345678901234567890123",
			wantError: ErrUserPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := NewPlainPassword(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if password.Value() != tt.input {
				t.Errorf("expected password value %q, got %q", tt.input, password.Value())
			}
		})
	}
}

func TestPlainPasswordEquals(t *testing.T) {
	password1, err := NewPlainPassword("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	password2, err := NewPlainPassword("password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	password3, err := NewPlainPassword("different123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !password1.Equals(password2) {
		t.Error("expected identical passwords to be equal")
	}

	if password1.Equals(password3) {
		t.Error("expected different passwords to not be equal")
	}
}
