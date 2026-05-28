package task

import "strings"

type Name struct {
	value string
}

func NewName(name string) (Name, error) {

	if name == "" {
		return Name{}, ErrTaskNameCannotBeEmpty
	}

	return Name{value: strings.TrimSpace(name)}, nil
}

func (n Name) String() string {
	return n.value
}
