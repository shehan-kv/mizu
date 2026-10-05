package contract

import "strings"

type Name struct {
	value string
}

func NewName(name string) (Name, error) {
	tName := strings.TrimSpace(name)

	if tName == "" {
		return Name{}, ErrContractNameCannotBeEmpty
	}

	return Name{value: tName}, nil
}

func (n Name) String() string {
	return n.value
}
