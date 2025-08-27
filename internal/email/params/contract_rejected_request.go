package params

type ContractRejectedRequest struct {
	ContractName    string
	ContractVersion string
	ContractText    string
	Signatures      []UserSignature
}
