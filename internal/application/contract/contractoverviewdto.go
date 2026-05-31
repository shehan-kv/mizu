package contract

import "time"

type ContractOverviewDTO struct {
	ID          string
	ProjectID   string
	Name        string
	Status      string
	Signatories []SignatoryDTO

	CreatedAt time.Time
	UpdatedAt time.Time
}
