package contract

import "testing"

func TestNewSignatoryStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect SignatoryStatus
	}{
		{
			name:   "pending",
			input:  "pending",
			expect: SignatoryStatusPending,
		},
		{
			name:   "signed",
			input:  "signed",
			expect: SignatoryStatusSigned,
		},
		{
			name:   "rejected",
			input:  "rejected",
			expect: SignatoryStatusRejected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewSignatoryStatus(tt.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if status != tt.expect {
				t.Fatalf("expected %q, got %q", tt.expect, status)
			}
		})
	}

	t.Run("rejects invalid status", func(t *testing.T) {
		_, err := NewSignatoryStatus("invalid")
		if err != ErrContractInvalidSignatoryStatus {
			t.Fatalf("expected %v, got %v", ErrContractInvalidSignatoryStatus, err)
		}
	})
}
