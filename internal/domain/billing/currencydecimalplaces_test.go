package billing

import "testing"

func TestNewCurrencyDecimals(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    CurrencyDecimals
		wantErr error
	}{
		{
			name:  "zero decimals",
			input: 0,
			want:  CurrencyDecimals(0),
		},
		{
			name:  "two decimals",
			input: 2,
			want:  CurrencyDecimals(2),
		},
		{
			name:  "positive decimals",
			input: 8,
			want:  CurrencyDecimals(8),
		},
		{
			name:    "negative decimals",
			input:   -1,
			wantErr: ErrBillingCurrencyDecimalsCannotBeNegative,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decimals, err := NewCurrencyDecimals(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if decimals != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, decimals)
			}

			if decimals.Int() != tt.input {
				t.Fatalf("expected Int() %d, got %d", tt.input, decimals.Int())
			}
		})
	}
}
