package aggregates

import "time"

type ContractRevisionWithUser struct {
	Id               int64
	ContractId       int64
	Title            string
	Description      string
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	Status           string
	ReqUserFirstName string
	ReqUserLastName  string
	ResUserFirstName *string
	ResUserLastName  *string
}
