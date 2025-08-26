package params

import "time"

type UserSignature struct {
	FirstName string
	LastName  string
	Email     string
	SignedAt  *time.Time
	Status    string
}

type ContractSignedRequest struct {
	ContractName    string
	ContractVersion string
	ContractText    string
	Signatures      []UserSignature
}
