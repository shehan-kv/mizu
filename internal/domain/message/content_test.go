package message

import (
	"strings"
	"testing"
)

func TestNewContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "valid content",
			input: "Hello, world!",
			want:  "Hello, world!",
		},
		{
			name:    "empty content",
			input:   "",
			wantErr: ErrMessageContentCannotBeEmpty,
		},
		{
			name:    "whitespace only",
			input:   "   \t\n  ",
			wantErr: ErrMessageContentCannotBeEmpty,
		},
		{
			name:  "content at maximum length",
			input: strings.Repeat("a", ContentMaxLength),
			want:  strings.Repeat("a", ContentMaxLength),
		},
		{
			name:    "content exceeds maximum length",
			input:   strings.Repeat("a", ContentMaxLength+1),
			wantErr: ErrMessageContentTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewContent(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if content.String() != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, content.String())
			}
		})
	}
}
