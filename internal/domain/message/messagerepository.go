package message

import (
	"context"
	"mizu/internal/domain/common"
)

type MessageRepository interface {
	Add(ctx context.Context, m *Message) error

	Get(ctx context.Context, id MessageID) (*Message, error)
	ListByChannel(ctx context.Context, cID ChannelID, p common.Page) ([]*Message, error)
	CountByChannel(ctx context.Context, cID ChannelID) (int, error)
}
