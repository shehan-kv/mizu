package aggregates

import "time"

type ContractSignature struct {
	FirstName string
	LastName  string
	Email     string
	SignedAt  *time.Time
	Status    string
}

type ContractUserSignatures struct {
	ContractName    string
	ContractVersion string
	ContractText    string
	Signatures      []ContractSignature
}
