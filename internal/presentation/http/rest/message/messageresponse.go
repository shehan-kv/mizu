package message

import "time"

type MessageResponse struct {
	ID        string                 `json:"id"`
	ChannelID string                 `json:"channelId"`
	IsSystem  bool                   `json:"isSystem"`
	Content   string                 `json:"content"`
	Sender    *MessageSenderResponse `json:"sender"`
	CreatedAt time.Time              `json:"createdAt"`
}
