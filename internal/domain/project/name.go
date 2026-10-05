package project

type Name struct {
	value string
}

func NewName(name string) (Name, error) {

	if name == "" {
		return Name{}, ErrProjectNameCannotBeEmpty
	}

	return Name{value: name}, nil
}

func (n Name) String() string {
	return n.value
}
