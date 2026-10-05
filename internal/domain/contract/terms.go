package contract

import "unicode/utf8"

type Terms struct {
	value string
}

func NewTerms(value string) (Terms, error) {
	if utf8.RuneCountInString(value) < 50 {
		return Terms{}, ErrContractTermsTooShort
	}

	return Terms{value: value}, nil
}

func (t Terms) String() string {
	return t.value
}
