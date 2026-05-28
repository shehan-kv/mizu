package contract

type ContractID string

func NewContractID(id string) (ContractID, error) {
	if id == "" {
		return "", ErrContractIDCannotBeEmpty
	}

	return ContractID(id), nil
}
func (c ContractID) String() string {
	return string(c)
}
