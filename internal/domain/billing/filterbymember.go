package billing

import (
	"mizu/internal/domain/iam"
)

type FilterByMember struct {
	MemberID  iam.UserID
	Keyword   *string
	IsInvoice *bool
	Status    *Status
}
