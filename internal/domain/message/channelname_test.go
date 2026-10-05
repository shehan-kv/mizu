package message

import (
	"testing"
)

func TestNewChannelName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "valid name",
			input: "general",
			want:  "general",
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: ErrChannelNameCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewChannelName(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if name.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, name.String())
			}
		})
	}
}
