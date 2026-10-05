package contract

import (
	"mizu/internal/domain/iam"
)

type FilterBySignatory struct {
	SignatoryID iam.UserID
	Keyword     *string
	Status      *Status
}
