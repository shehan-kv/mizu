package message

import (
	"time"
)

type FileDTO struct {
	ID         string
	ChannelID  string
	User       MessageSenderDTO
	Name       string
	MimeType   string
	Size       int64
	UploadedAt time.Time
}
