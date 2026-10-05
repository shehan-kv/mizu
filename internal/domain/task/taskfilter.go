package task

import "mizu/internal/domain/project"

type TaskFilter struct {
	ProjectID project.ProjectID
	Keyword   *string
	Status    *Status
	Priority  *Priority
}
