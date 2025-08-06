package user

import (
	"net/mail"
	"strings"
)

// Represents a user create create request
type UserCreateRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Title     string `json:"title"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *UserCreateRequest) Validate() bool {
	r.format()

	if len(r.FirstName) == 0 {
		return false
	}

	if len(r.LastName) == 0 {
		return false
	}

	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return false
	}

	if len(r.Role) == 0 {
		return false
	}

	return true
}

func (r *UserCreateRequest) format() {
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.Title = strings.TrimSpace(r.Title)
	r.Role = strings.ToLower(strings.TrimSpace(r.Role))
}
