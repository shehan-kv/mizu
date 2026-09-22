package billing

import "testing"

func TestNewCurrencyCode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CurrencyCode
		wantErr error
	}{
		{
			name:  "uppercase code",
			input: "USD",
			want:  CurrencyCode("USD"),
		},
		{
			name:  "lowercase code is normalized",
			input: "usd",
			want:  CurrencyCode("USD"),
		},
		{
			name:  "mixed case code is normalized",
			input: "uSd",
			want:  CurrencyCode("USD"),
		},
		{
			name:    "empty code",
			input:   "",
			wantErr: ErrBillingCurrencyCodeCannotBeEmpty,
		},
		{
			name:    "code too short",
			input:   "US",
			wantErr: ErrBillingInvalidCurrencyCode,
		},
		{
			name:    "code too long",
			input:   "USDX",
			wantErr: ErrBillingInvalidCurrencyCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := NewCurrencyCode(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if code != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, code)
			}

			if code.String() != string(tt.want) {
				t.Fatalf("expected String() %q, got %q", tt.want, code.String())
			}
		})
	}
}
