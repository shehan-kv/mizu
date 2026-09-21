package iam

import "testing"

func TestNewImageName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ImageName
		wantError error
	}{
		{
			name:  "valid image name",
			input: "profile.png",
			want:  ImageName("profile.png"),
		},
		{
			name:      "empty image name",
			input:     "",
			wantError: ErrUserImageNameCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageName, err := NewImageName(tt.input)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if imageName != tt.want {
				t.Errorf("expected %q, got %q", tt.want, imageName)
			}

			if imageName.String() != tt.input {
				t.Errorf("expected String() to return %q, got %q", tt.input, imageName.String())
			}
		})
	}
}
