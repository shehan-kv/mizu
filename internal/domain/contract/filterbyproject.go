package contract

import "mizu/internal/domain/project"

type FilterByProject struct {
	ProjectID project.ProjectID
	Keyword   *string
	Status    *Status
}
