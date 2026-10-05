package billing

import "testing"

func TestNewCurrency(t *testing.T) {
	name, err := NewCurrencyName("US Dollar")
	if err != nil {
		t.Fatal(err)
	}

	symbol, err := NewCurrencySymbol("$")
	if err != nil {
		t.Fatal(err)
	}

	code, err := NewCurrencyCode("USD")
	if err != nil {
		t.Fatal(err)
	}

	decimals, err := NewCurrencyDecimals(2)
	if err != nil {
		t.Fatal(err)
	}

	currency := NewCurrency(name, symbol, code, decimals)

	if currency.Name() != name {
		t.Fatalf("expected name %q, got %q", name, currency.Name())
	}

	if currency.Symbol() != symbol {
		t.Fatalf("expected symbol %q, got %q", symbol, currency.Symbol())
	}

	if currency.Code() != code {
		t.Fatalf("expected code %q, got %q", code, currency.Code())
	}

	if currency.DecimalPlaces() != decimals {
		t.Fatalf("expected decimal places %d, got %d", decimals, currency.DecimalPlaces())
	}
}
