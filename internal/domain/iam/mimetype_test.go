package iam

import "testing"

func TestNewMimeType(t *testing.T) {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
		"image/avif",
	}

	for _, input := range validTypes {
		t.Run(input, func(t *testing.T) {
			mimeType, err := NewMimeType(input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if mimeType.String() != input {
				t.Errorf("expected %q, got %q", input, mimeType.String())
			}
		})
	}
}

func TestNewMimeTypeRejectsInvalidTypes(t *testing.T) {
	invalidTypes := []string{
		"",
		"text/plain",
		"application/json",
		"image/bmp",
		"IMAGE/PNG",
	}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			mimeType, err := NewMimeType(input)

			if err != ErrUserInvalidImageMimeType {
				t.Fatalf(
					"expected error %v, got %v",
					ErrUserInvalidImageMimeType,
					err,
				)
			}

			if mimeType != "" {
				t.Errorf("expected empty MIME type, got %q", mimeType)
			}
		})
	}
}
