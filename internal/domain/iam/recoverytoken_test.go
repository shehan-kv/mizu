package iam

import "testing"

func TestNewRecoveryToken(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      RecoveryToken
		wantError error
	}{
		{
			name:  "valid token",
			input: "abc123",
			want:  RecoveryToken("abc123"),
		},
		{
			name:      "empty token",
			input:     "",
			wantError: ErrRecoveryTokenCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewRecoveryToken(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.want {
				t.Errorf("expected token %q, got %q", tt.want, token)
			}

			if token.String() != tt.input {
				t.Errorf("expected String() to return %q, got %q", tt.input, token.String())
			}
		})
	}
}
