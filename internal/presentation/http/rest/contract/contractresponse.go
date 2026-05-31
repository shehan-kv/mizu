package contract

import "time"

type ContractResponse struct {
	ID          string              `json:"id"`
	ProjectID   string              `json:"projectId"`
	Name        string              `json:"name"`
	Status      string              `json:"status"`
	Terms       string              `json:"terms"`
	Signatories []SignatoryResponse `json:"signatories"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}
