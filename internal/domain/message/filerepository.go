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
	ListByChannel(ctx context.Context, f FileFilter, p common.Page) ([]*File, error)

	CountByChannel(ctx context.Context, f FileFilter) (int, error)
}
