package task

import "testing"

func TestNewTaskID(t *testing.T) {
	t.Run("creates task ID", func(t *testing.T) {
		id, err := NewTaskID("task-1")
		if err != nil {
			t.Fatalf("NewTaskID() error = %v", err)
		}

		if id.String() != "task-1" {
			t.Fatalf("expected task-1, got %q", id.String())
		}
	})

	t.Run("rejects empty ID", func(t *testing.T) {
		_, err := NewTaskID("")
		if err != ErrTaskIDCannotBeEmpty {
			t.Fatalf("expected ErrTaskIDCannotBeEmpty, got %v", err)
		}
	})
}
