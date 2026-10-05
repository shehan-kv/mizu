package task

import "testing"

func TestNewStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Status
		wantErr error
	}{
		{
			name:  "backlog",
			input: "backlog",
			want:  StatusBacklog,
		},
		{
			name:  "in progress",
			input: "in-progress",
			want:  StatusInProgress,
		},
		{
			name:  "completed",
			input: "completed",
			want:  StatusCompleted,
		},
		{
			name:    "invalid status",
			input:   "cancelled",
			wantErr: ErrTaskInvalidStatus,
		},
		{
			name:    "empty status",
			input:   "",
			wantErr: ErrTaskInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStatus(tt.input)

			if err != tt.wantErr {
				t.Fatalf("NewStatus() error = %v, want %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Fatalf("NewStatus() = %v, want %v", got, tt.want)
			}

			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}
