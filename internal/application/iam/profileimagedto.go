package iam

import "io"

type ProfileImageDTO struct {
	Name     string
	MimeType string
	Reader   io.ReadCloser
}
