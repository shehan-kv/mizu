package project

import "testing"

func TestNewName(t *testing.T) {
	t.Run("creates project name", func(t *testing.T) {
		name, err := NewName("MizuPM")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if name.String() != "MizuPM" {
			t.Fatalf("expected %q, got %q", "MizuPM", name.String())
		}
	})

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := NewName("")

		if err != ErrProjectNameCannotBeEmpty {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectNameCannotBeEmpty,
				err,
			)
		}
	})
}
