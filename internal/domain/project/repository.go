package project

import (
	"context"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
)

type Repository interface {
	Add(ctx context.Context, p *Project) error

	Get(ctx context.Context, id ProjectID) (*Project, error)

	List(ctx context.Context, f Filter, p common.Page) ([]*Project, error)
	ListByIDs(ctx context.Context, pIDs []ProjectID) ([]*Project, error)
	ListByMember(ctx context.Context, mID iam.UserID) ([]*Project, error)
	ListCreatedPerDay(ctx context.Context, mID iam.UserID) ([]Metric, error)

	Count(ctx context.Context, f Filter) (int, error)

	Save(ctx context.Context, p *Project) error
	SaveAll(ctx context.Context, p []*Project) error

	Remove(ctx context.Context, p *Project) error
}
