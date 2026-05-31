package contract

type CreateContractRequest struct {
	Name         string   `json:"name"`
	Terms        string   `json:"terms"`
	SignatoryIDs []string `json:"signatoryIds"`
}
