package task

import (
	"context"
	"mizu/internal/domain/common"
	"mizu/internal/domain/project"
)

type Repository interface {
	Add(ctx context.Context, t *Task) error

	Get(ctx context.Context, tID TaskID) (*Task, error)
	GetStatsByProject(ctx context.Context, pID project.ProjectID) (Stats, error)

	List(ctx context.Context, f TaskFilter, p common.Page) ([]*Task, error)
	ListCompletedPerDay(ctx context.Context, pID project.ProjectID) ([]Metric, error)
	ListStatsByProjects(ctx context.Context, pIDs []project.ProjectID) (map[project.ProjectID]Stats, error)

	Count(ctx context.Context, f TaskFilter) (int, error)

	Save(ctx context.Context, t *Task) error

	Remove(ctx context.Context, t *Task) error
}
