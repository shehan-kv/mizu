package billing

import "testing"

func TestNewQty(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "positive integer",
			input: "5",
			want:  "5",
		},
		{
			name:  "positive decimal",
			input: "2.5",
			want:  "2.5",
		},
		{
			name:    "zero",
			input:   "0",
			wantErr: ErrBillingQtyMustBePositive,
		},
		{
			name:    "negative quantity",
			input:   "-1",
			wantErr: ErrBillingQtyMustBePositive,
		},
		{
			name:    "invalid decimal",
			input:   "invalid",
			wantErr: ErrBillingInvalidDecimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qty, err := NewQty(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if qty.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, qty.String())
			}

			if qty.ToDecimal().String() != tt.want {
				t.Fatalf("expected decimal %q, got %q", tt.want, qty.ToDecimal().String())
			}
		})
	}
}
