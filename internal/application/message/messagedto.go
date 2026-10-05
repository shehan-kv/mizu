package message

import (
	"time"
)

type MessageDTO struct {
	ID        string
	ChannelID string
	IsSystem  bool
	Content   string
	Sender    *MessageSenderDTO
	CreatedAt time.Time
}
