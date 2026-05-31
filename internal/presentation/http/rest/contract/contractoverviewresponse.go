package contract

import "time"

type ContractOverviewResponse struct {
	ID          string              `json:"id"`
	ProjectID   string              `json:"projectId"`
	Name        string              `json:"name"`
	Status      string              `json:"status"`
	Signatories []SignatoryResponse `json:"signatories"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}
