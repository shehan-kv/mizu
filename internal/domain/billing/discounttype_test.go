package billing

import "testing"

func TestNewDiscountType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DiscountType
		wantErr error
	}{
		{
			name:  "fixed",
			input: "fixed",
			want:  DiscountTypeFixed,
		},
		{
			name:  "percentage",
			input: "percentage",
			want:  DiscountTypePercentage,
		},
		{
			name:    "invalid type",
			input:   "invalid",
			wantErr: ErrBillingInvalidDiscountType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dType, err := NewDiscountType(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dType != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, dType)
			}

			if dType.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, dType.String())
			}
		})
	}
}
