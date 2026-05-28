package message

import (
	"context"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type ChannelRepository interface {
	Add(ctx context.Context, c *Channel) error

	Get(ctx context.Context, cID ChannelID) (*Channel, error)
	ListByProject(ctx context.Context, pID project.ProjectID) ([]*Channel, error)
	ListByMember(ctx context.Context, mID iam.UserID) ([]*Channel, error)

	Save(ctx context.Context, c *Channel) error
}
