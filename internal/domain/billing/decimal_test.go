package billing

import "testing"

func mustDecimal(t *testing.T, value string) Decimal {
	t.Helper()

	decimal, err := NewDecimal(value)
	if err != nil {
		t.Fatalf("failed to create decimal %q: %v", value, err)
	}

	return decimal
}

func TestNewDecimal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "integer",
			input: "100",
			want:  "100",
		},
		{
			name:  "decimal",
			input: "12.50",
			want:  "12.50",
		},
		{
			name:  "negative decimal",
			input: "-12.50",
			want:  "-12.50",
		},
		{
			name:    "invalid decimal",
			input:   "not-a-number",
			wantErr: ErrBillingInvalidDecimal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decimal, err := NewDecimal(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if decimal.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, decimal.String())
			}
		})
	}
}

func TestDecimalAdd(t *testing.T) {
	a := mustDecimal(t, "10.50")
	b := mustDecimal(t, "2.25")
	expected := mustDecimal(t, "12.75")

	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(expected) {
		t.Fatalf("expected 12.75, got %s", result.String())
	}
}

func TestDecimalSubtract(t *testing.T) {
	a := mustDecimal(t, "10.50")
	b := mustDecimal(t, "2.25")
	expected := mustDecimal(t, "8.25")

	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(expected) {
		t.Fatalf("expected 8.25, got %s", result.String())
	}
}

func TestDecimalMultiply(t *testing.T) {
	a := mustDecimal(t, "10.50")
	b := mustDecimal(t, "2.5")
	expected := mustDecimal(t, "26.25")

	result, err := a.Multiply(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(expected) {
		t.Fatalf("expected %s, got %s", expected.String(), result.String())
	}
}

func TestDecimalDivide(t *testing.T) {
	a := mustDecimal(t, "10")
	b := mustDecimal(t, "4")
	expected := mustDecimal(t, "2.5")

	result, err := a.Divide(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(expected) {
		t.Fatalf("expected %s, got %s", expected.String(), result.String())
	}
}

func TestDecimalDivideByZero(t *testing.T) {
	a := mustDecimal(t, "10")
	zero := mustDecimal(t, "0")

	_, err := a.Divide(zero)

	if err != ErrBillingDivisionFailed {
		t.Fatalf("expected %v, got %v", ErrBillingDivisionFailed, err)
	}
}

func TestDecimalGreaterThan(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "greater",
			a:    "10",
			b:    "5",
			want: true,
		},
		{
			name: "equal",
			a:    "10",
			b:    "10",
			want: false,
		},
		{
			name: "less",
			a:    "5",
			b:    "10",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mustDecimal(t, tt.a)
			b := mustDecimal(t, tt.b)

			if got := a.GreaterThan(b); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDecimalEquals(t *testing.T) {
	a := mustDecimal(t, "10.00")
	b := mustDecimal(t, "10.00")
	c := mustDecimal(t, "10.01")

	if !a.Equals(b) {
		t.Fatal("expected equal decimals to be equal")
	}

	if a.Equals(c) {
		t.Fatal("expected different decimals not to be equal")
	}
}
