package billing

import "testing"

func TestNewTaxType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TaxType
		wantErr error
	}{
		{
			name:  "fixed",
			input: "fixed",
			want:  TaxTypeFixed,
		},
		{
			name:  "percentage",
			input: "percentage",
			want:  TaxTypePercentage,
		},
		{
			name:    "invalid type",
			input:   "invalid",
			wantErr: ErrBillingInvalidTaxType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxType, err := NewTaxType(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if taxType != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, taxType)
			}

			if taxType.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, taxType.String())
			}
		})
	}
}
