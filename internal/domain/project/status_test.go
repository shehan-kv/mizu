package project

import "testing"

func TestNewStatus(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect Status
	}{
		{
			name:   "started",
			input:  "started",
			expect: StatusStarted,
		},
		{
			name:   "paused",
			input:  "paused",
			expect: StatusPaused,
		},
		{
			name:   "cancelled",
			input:  "cancelled",
			expect: StatusCancelled,
		},
		{
			name:   "completed",
			input:  "completed",
			expect: StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewStatus(tt.input)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if status != tt.expect {
				t.Fatalf(
					"expected %q, got %q",
					tt.expect,
					status,
				)
			}
		})
	}

	t.Run("rejects invalid status", func(t *testing.T) {
		_, err := NewStatus("invalid")

		if err != ErrProjectInvalidStatus {
			t.Fatalf(
				"expected %v, got %v",
				ErrProjectInvalidStatus,
				err,
			)
		}
	})
}
