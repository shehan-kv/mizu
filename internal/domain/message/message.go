package message

import (
	"mizu/internal/domain/iam"
	"time"
)

type Message struct {
	id        MessageID
	channelID ChannelID
	senderID  iam.UserID
	isSystem  bool
	content   Content
	createdAt time.Time
}

func NewUserMessage(
	id MessageID,
	channelID ChannelID,
	senderID iam.UserID,
	content Content,
	now time.Time,
) (*Message, error) {

	if senderID == iam.SystemUserID {
		return nil, ErrMessageInvalidSender
	}

	return &Message{
		id:        id,
		channelID: channelID,
		senderID:  senderID,
		isSystem:  false,
		content:   content,
		createdAt: now,
	}, nil
}

func NewSystemMessage(
	id MessageID,
	channelID ChannelID,
	content Content,
	now time.Time,
) *Message {

	return &Message{
		id:        id,
		channelID: channelID,
		senderID:  iam.SystemUserID,
		isSystem:  true,
		content:   content,
		createdAt: now,
	}
}

func RestoreMessage(
	id MessageID,
	channelID ChannelID,
	senderID iam.UserID,
	isSystem bool,
	content Content,
	createdAt time.Time,
) *Message {

	return &Message{
		id:        id,
		channelID: channelID,
		senderID:  senderID,
		isSystem:  isSystem,
		content:   content,
		createdAt: createdAt,
	}
}

func (m Message) ID() MessageID {
	return m.id
}

func (m Message) ChannelID() ChannelID {
	return m.channelID
}

func (m Message) SenderID() iam.UserID {
	return m.senderID
}

func (m Message) IsSystemMessage() bool {
	return m.isSystem
}

func (m Message) IsUserMessage() bool {
	return !m.isSystem
}

func (m Message) Content() Content {
	return m.content
}

func (m Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m Message) Equals(other Message) bool {
	return m.id == other.id
}
