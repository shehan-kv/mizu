package message

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

var (
	EventTypeFileCreated common.EventType = "message.file.created"
)

type FileCreatedEvent struct {
	FileID       FileID
	ChannelID    ChannelID
	UserID       iam.UserID
	OriginalName FileName
	StorageKey   string
	MimeType     string
	Size         int64
	OccurredAt   time.Time
}

func (e FileCreatedEvent) EventType() common.EventType {
	return EventTypeFileCreated
}
func (e FileCreatedEvent) EventScope() common.EventScope {
	return common.EventScopeInternal
}
