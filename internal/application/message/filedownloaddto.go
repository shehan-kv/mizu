package message

import (
	"io"
	"time"
)

type FileDownloadDTO struct {
	Name       string
	MimeType   string
	Size       int64
	Reader     io.ReadCloser
	UploadedAt time.Time
}
