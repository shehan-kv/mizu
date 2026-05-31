package contract

import "time"

type ContractDTO struct {
	ID          string
	ProjectID   string
	Name        string
	Status      string
	Terms       string
	Signatories []SignatoryDTO

	CreatedAt time.Time
	UpdatedAt time.Time
}
