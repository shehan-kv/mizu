package iam

import "io"

type ProfileImageParams struct {
	FileName string
	MimeType string
	Size     int64
	Reader   io.Reader
}
