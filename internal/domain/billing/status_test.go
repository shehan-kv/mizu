package billing

import "testing"

func TestNewStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Status
		wantErr error
	}{
		{
			name:  "paid",
			input: "paid",
			want:  StatusPaid,
		},
		{
			name:  "pending",
			input: "pending",
			want:  StatusPending,
		},
		{
			name:  "accepted",
			input: "accepted",
			want:  StatusAccepted,
		},
		{
			name:  "rejected",
			input: "rejected",
			want:  StatusRejected,
		},
		{
			name:  "cancelled",
			input: "cancelled",
			want:  StatusCancelled,
		},
		{
			name:    "invalid status",
			input:   "invalid",
			wantErr: ErrBillingInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewStatus(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if status != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, status)
			}

			if status.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, status.String())
			}
		})
	}
}
