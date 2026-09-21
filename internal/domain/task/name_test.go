package task

import "testing"

func TestNewName(t *testing.T) {
	t.Run("creates name", func(t *testing.T) {
		name, err := NewName("Implement authentication")
		if err != nil {
			t.Fatalf("NewName() error = %v", err)
		}

		if name.String() != "Implement authentication" {
			t.Fatalf("expected Implement authentication, got %q", name.String())
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		name, err := NewName("  Implement authentication  ")
		if err != nil {
			t.Fatalf("NewName() error = %v", err)
		}

		if name.String() != "Implement authentication" {
			t.Fatalf("expected trimmed name, got %q", name.String())
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := NewName("")
		if err != ErrTaskNameCannotBeEmpty {
			t.Fatalf("expected ErrTaskNameCannotBeEmpty, got %v", err)
		}
	})
}
