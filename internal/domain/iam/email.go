package iam

import (
	"net/mail"
)

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, ErrUserEmailCannotBeEmpty
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, ErrUserInvalidEmailAddress
	}

	return Email{value: address.Address}, nil
}

func (e Email) String() string {
	return e.value
}
