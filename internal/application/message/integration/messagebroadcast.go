package integration

import (
	"mizu/internal/application/eventbus"
	"time"
)

const EventMessageBroadcast eventbus.EventType = "integration.messaging.broadcast"

type MessageBroadcast struct {
	MessageID       string    `json:"messageId"`
	ChannelID       string    `json:"channelId"`
	To              []string  `json:"to"`
	SenderID        string    `json:"senderId"`
	SenderFirstName *string   `json:"senderFirstName"`
	SenderLastName  *string   `json:"senderLastName"`
	SenderImage     *string   `json:"senderImage"`
	SenderTitle     *string   `json:"senderTitle"`
	SenderRole      *string   `json:"senderRole"`
	IsSystem        bool      `json:"isSystem"`
	Content         string    `json:"content"`
	OccurredAt      time.Time `json:"occurredAt"`
}

func (e MessageBroadcast) EventType() eventbus.EventType {
	return EventMessageBroadcast
}
