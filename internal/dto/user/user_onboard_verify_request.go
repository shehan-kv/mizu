package user

// Represents a user onboard verify request
type UserOnboardVerifyRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *UserOnboardVerifyRequest) Validate() bool {

	if len(r.Password) < 8 {
		return false
	}

	if r.Password != r.ConfirmPassword {
		return false
	}

	return true
}
