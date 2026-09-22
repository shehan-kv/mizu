package billing

import "testing"

func TestNewCurrencySymbol(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CurrencySymbol
		wantErr error
	}{
		{
			name:  "valid symbol",
			input: "$",
			want:  CurrencySymbol("$"),
		},
		{
			name:  "multi-character symbol",
			input: "Rs.",
			want:  CurrencySymbol("Rs."),
		},
		{
			name:    "empty symbol",
			input:   "",
			wantErr: ErrBillingCurrencySymbolCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symbol, err := NewCurrencySymbol(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if symbol != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, symbol)
			}

			if symbol.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, symbol.String())
			}
		})
	}
}
