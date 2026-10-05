package mailer

import "time"

type ContractSignatory struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Title     *string
	Role      string

	Status string

	UpdatedAt time.Time
}
