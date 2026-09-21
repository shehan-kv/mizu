package iam

import "testing"

func TestNewImage(t *testing.T) {
	name := ImageName("avatar.png")
	mimeType := MimeType("image/png")

	image := NewImage(name, mimeType)

	if image.Name() != name {
		t.Errorf("expected image name %q, got %q", name, image.Name())
	}

	if image.MimeType() != mimeType {
		t.Errorf("expected MIME type %q, got %q", mimeType, image.MimeType())
	}
}
