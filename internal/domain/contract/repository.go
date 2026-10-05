package contract

import (
	"context"
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
)

type Repository interface {
	Add(ctx context.Context, c *Contract) error

	Get(ctx context.Context, id ContractID) (*Contract, error)
	GetStatsByProject(ctx context.Context, pID project.ProjectID) (Stats, error)

	ListByProject(ctx context.Context, f FilterByProject, p common.Page) ([]*Contract, error)
	ListBySignatory(ctx context.Context, f FilterBySignatory, p common.Page) ([]*Contract, error)
	ListStatsByProjects(ctx context.Context, pIDs []project.ProjectID) (map[project.ProjectID]Stats, error)

	CountByProject(ctx context.Context, f FilterByProject) (int, error)
	CountBySignatory(ctx context.Context, f FilterBySignatory) (int, error)

	Save(ctx context.Context, c *Contract) error
}
