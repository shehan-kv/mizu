package billing

import "testing"

func TestNewCurrencyName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CurrencyName
		wantErr error
	}{
		{
			name:  "valid name",
			input: "US Dollar",
			want:  CurrencyName("US Dollar"),
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: ErrBillingCurrencyNameCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewCurrencyName(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if name != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, name)
			}

			if name.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, name.String())
			}
		})
	}
}
