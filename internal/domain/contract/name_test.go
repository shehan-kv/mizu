package contract

import "testing"

func TestNewName(t *testing.T) {
	t.Run("creates name", func(t *testing.T) {
		name, err := NewName("Employment Contract")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if name.String() != "Employment Contract" {
			t.Fatalf("expected %q, got %q", "Employment Contract", name.String())
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		name, err := NewName("  Employment Contract  ")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if name.String() != "Employment Contract" {
			t.Fatalf("expected trimmed name, got %q", name.String())
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := NewName("")
		if err != ErrContractNameCannotBeEmpty {
			t.Fatalf("expected %v, got %v", ErrContractNameCannotBeEmpty, err)
		}
	})
}
