package message

import (
	"context"
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
)

type FileRepository interface {
	Add(ctx context.Context, f *File) error

	Get(ctx context.Context, id FileID) (*File, error)
	GetStatsByProject(ctx context.Context, pID project.ProjectID) (Stats, error)
	ListByChannel(ctx context.Context, f ChannelFileFilter, p common.Page) ([]*File, error)
	ListByProject(ctx context.Context, f ProjectFileFilter, p common.Page) ([]*File, error)

	CountByChannel(ctx context.Context, f ChannelFileFilter) (int, error)
	CountByProject(ctx context.Context, f ProjectFileFilter) (int, error)
}
