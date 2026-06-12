package contract

import "time"

type ContractOverviewDTO struct {
	ID                    string
	ProjectID             string
	Name                  string
	Status                string
	MemberSignatoryStatus string
	Signatories           []SignatoryDTO

	CreatedAt time.Time
	UpdatedAt time.Time
}
