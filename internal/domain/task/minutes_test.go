package task

import "testing"

func TestNewMinutes(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    Minutes
		wantErr error
	}{
		{
			name:  "positive minutes",
			input: 60,
			want:  Minutes(60),
		},
		{
			name:    "zero minutes",
			input:   0,
			wantErr: ErrTaskMinutesMustBePositive,
		},
		{
			name:    "negative minutes",
			input:   -1,
			wantErr: ErrTaskMinutesMustBePositive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMinutes(tt.input)

			if err != tt.wantErr {
				t.Fatalf("NewMinutes() error = %v, want %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Fatalf("NewMinutes() = %v, want %v", got, tt.want)
			}

			if got.Int() != tt.input {
				t.Fatalf("Int() = %d, want %d", got.Int(), tt.input)
			}
		})
	}
}
