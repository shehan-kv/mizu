package message

import "time"

type FileResponse struct {
	ID         string                `json:"id"`
	ChannelID  string                `json:"channelId"`
	User       MessageSenderResponse `json:"user"`
	Name       string                `json:"name"`
	MimeType   string                `json:"mimeType"`
	Size       int64                 `json:"size"`
	UploadedAt time.Time             `json:"uploadedAt"`
}
