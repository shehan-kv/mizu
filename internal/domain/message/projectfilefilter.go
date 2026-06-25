package message

import (
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
)

type ProjectFileFilter struct {
	MemberID  iam.UserID
	ProjectID project.ProjectID
	Keyword   *string
}
