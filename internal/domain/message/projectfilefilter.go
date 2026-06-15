package message

import "mizu/internal/domain/project"

type ProjectFileFilter struct {
	ProjectID project.ProjectID
	Keyword   *string
}
