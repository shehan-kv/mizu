package contract

type CreateContractParams struct {
	ActorID      string
	ProjectID    string
	Name         string
	Terms        string
	SignatoryIDs []string
}
