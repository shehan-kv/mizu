package iam

import "testing"

func TestNewVerificationID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    VerificationID
		wantErr error
	}{
		{
			name:  "valid ID",
			input: "verification-1",
			want:  VerificationID("verification-1"),
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: ErrVerificationIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewVerificationID(tt.input)

			if err != tt.wantErr {
				t.Fatalf("NewVerificationID() error = %v, want %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Fatalf("NewVerificationID() = %v, want %v", got, tt.want)
			}

			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}
