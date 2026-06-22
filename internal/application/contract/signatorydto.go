package contract

import "time"

type SignatoryDTO struct {
	ID        string
	FirstName string
	LastName  string
	Title     *string
	Role      string
	HasImage  bool
	Status    string
	UpdatedAt time.Time
}
