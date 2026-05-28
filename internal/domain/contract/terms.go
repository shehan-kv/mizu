package contract

type Terms struct {
	value string
}

func NewTerms(value string) (Terms, error) {
	if len(value) < 50 {
		return Terms{}, ErrContractTermsTooShort
	}

	return Terms{value: value}, nil
}

func (t Terms) String() string {
	return t.value
}
