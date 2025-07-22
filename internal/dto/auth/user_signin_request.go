package dto

import (
	"net/mail"
	"strings"
)

// Represents a user sign in request
type UserSignInRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (usr *UserSignInRequest) Validate() bool {
	usr.format()

	_, err := mail.ParseAddress(usr.Email)
	if err != nil {
		return false
	}

	if len(usr.Password) < 8 {
		return false
	}

	return true
}

func (usr *UserSignInRequest) format() {
	usr.Email = strings.ToLower(strings.TrimSpace(usr.Email))
}
