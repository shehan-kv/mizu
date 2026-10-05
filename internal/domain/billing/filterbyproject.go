package billing

import "mizu/internal/domain/project"

type FilterByProject struct {
	ProjectID project.ProjectID
	Keyword   *string
	IsInvoice *bool
	Status    *Status
}
