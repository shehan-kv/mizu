package task

import "testing"

func TestNewPriority(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Priority
		wantErr error
	}{
		{
			name:  "high",
			input: "high",
			want:  PriorityHigh,
		},
		{
			name:  "medium",
			input: "medium",
			want:  PriorityMedium,
		},
		{
			name:  "low",
			input: "low",
			want:  PriorityLow,
		},
		{
			name:    "invalid priority",
			input:   "urgent",
			wantErr: ErrTaskInvalidPriority,
		},
		{
			name:    "empty priority",
			input:   "",
			wantErr: ErrTaskInvalidPriority,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPriority(tt.input)

			if err != tt.wantErr {
				t.Fatalf("NewPriority() error = %v, want %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Fatalf("NewPriority() = %v, want %v", got, tt.want)
			}

			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
		})
	}
}
