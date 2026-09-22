package message

import "testing"

func TestNewChannelID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ChannelID
		wantErr error
	}{
		{
			name:  "valid ID",
			input: "channel-1",
			want:  ChannelID("channel-1"),
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: ErrChannelIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewChannelID(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if id != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, id)
			}

			if id.String() != tt.input {
				t.Fatalf("expected String() to return %q, got %q", tt.input, id.String())
			}
		})
	}
}

func TestNewMessageID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MessageID
		wantErr error
	}{
		{
			name:  "valid ID",
			input: "message-1",
			want:  MessageID("message-1"),
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: ErrMessageIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewMessageID(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if id != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, id)
			}

			if id.String() != tt.input {
				t.Fatalf("expected String() to return %q, got %q", tt.input, id.String())
			}
		})
	}
}

func TestNewFileID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    FileID
		wantErr error
	}{
		{
			name:  "valid ID",
			input: "file-1",
			want:  FileID("file-1"),
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: ErrFileIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewFileID(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if id != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, id)
			}

			if id.String() != tt.input {
				t.Fatalf("expected String() to return %q, got %q", tt.input, id.String())
			}
		})
	}
}
