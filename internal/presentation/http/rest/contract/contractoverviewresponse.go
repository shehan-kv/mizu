package contract

import "time"

type ContractOverviewResponse struct {
	ID                    string              `json:"id"`
	ProjectID             string              `json:"projectId"`
	Name                  string              `json:"name"`
	Status                string              `json:"status"`
	MemberSignatoryStatus string              `json:"memberSignatoryStatus"`
	Signatories           []SignatoryResponse `json:"signatories"`
	CreatedAt             time.Time           `json:"createdAt"`
	UpdatedAt             time.Time           `json:"updatedAt"`
}
