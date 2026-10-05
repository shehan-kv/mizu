package project

import "mizu/internal/domain/iam"

type Filter struct {
	Keyword  *string
	Status   *Status
	MemberID iam.UserID
}
