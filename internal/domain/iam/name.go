package iam

type Name struct {
	firstName string
	lastName  string
}

func NewName(firstName string, lastName string) (Name, error) {

	if firstName == "" {
		return Name{}, ErrUserFirstNameCannotBeEmpty
	}

	if lastName == "" {
		return Name{}, ErrUserLastNameCannotBeEmpty
	}

	name := Name{
		firstName: firstName,
		lastName:  lastName,
	}

	return name, nil
}

func (n Name) FirstName() string {
	return n.firstName
}

func (n Name) LastName() string {
	return n.lastName
}

func (n Name) Equals(name Name) bool {
	return n == name
}
