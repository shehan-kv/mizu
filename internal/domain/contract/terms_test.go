package contract

import "testing"

func TestNewTerms(t *testing.T) {
	t.Run("accepts terms with minimum length", func(t *testing.T) {
		value := "12345678901234567890123456789012345678901234567890"

		terms, err := NewTerms(value)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if terms.String() != value {
			t.Fatalf("expected %q, got %q", value, terms.String())
		}
	})

	t.Run("rejects terms shorter than minimum length", func(t *testing.T) {
		_, err := NewTerms("too short")
		if err != ErrContractTermsTooShort {
			t.Fatalf("expected %v, got %v", ErrContractTermsTooShort, err)
		}
	})
}
