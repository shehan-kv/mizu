package message

import (
	"context"
)

type MessageRepository interface {
	Add(ctx context.Context, m *Message) error

	Get(ctx context.Context, id MessageID) (*Message, error)
	ListByChannel(
		ctx context.Context,
		cID ChannelID,
		before *MessageID,
		limit int,
	) ([]*Message, error)
}
