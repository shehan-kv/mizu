package project

import "testing"

func TestNewProjectID(t *testing.T) {
	t.Run("creates project ID", func(t *testing.T) {
		id, err := NewProjectID("project-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if id.String() != "project-123" {
			t.Fatalf("expected %q, got %q", "project-123", id.String())
		}
	})

	t.Run("rejects empty ID", func(t *testing.T) {
		_, err := NewProjectID("")

		if err != ErrProjectIDCannotBeEmpty {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectIDCannotBeEmpty,
				err,
			)
		}
	})
}
