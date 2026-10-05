package billing

import (
	"context"
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type Repository interface {
	Add(ctx context.Context, i *Invoice) error

	Get(ctx context.Context, id InvoiceID) (*Invoice, error)
	GetCurrencyByCode(ctx context.Context, c CurrencyCode) (Currency, error)
	GetBillingOverviewByMember(ctx context.Context, mID iam.UserID) (BillingOverview, error)
	GetStatsByProject(ctx context.Context, pID project.ProjectID) (Stats, error)

	ListByMember(ctx context.Context, f FilterByMember, p common.Page) ([]*Invoice, error)
	ListByProject(ctx context.Context, f FilterByProject, p common.Page) ([]*Invoice, error)
	ListMonthlyPaidCountByProject(ctx context.Context, pID project.ProjectID) ([]Metric, error)
	ListMonthlyPaidCountByMember(ctx context.Context, mID iam.UserID) ([]Metric, error)
	ListStatsByProjects(ctx context.Context, pIDs []project.ProjectID) (map[project.ProjectID]Stats, error)

	CountByProject(ctx context.Context, f FilterByProject) (int, error)
	CountByMember(ctx context.Context, f FilterByMember) (int, error)

	Save(ctx context.Context, i *Invoice) error
}
