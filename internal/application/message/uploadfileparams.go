package message

import "io"

type UploadFileParams struct {
	ActorID   string
	ChannelID string
	FileName  string
	MimeType  string
	Size      int64
	Reader    io.Reader
}
