package aggregates

import "time"

type ContractSignature struct {
	Id        int64
	FirstName string
	LastName  string
	Email     string
	SignedAt  *time.Time
	Image     *string
	Status    string
}

type ContractUserSignatures struct {
	ContractName    string
	ContractVersion string
	ContractText    string
	Signatures      []ContractSignature
}
