package dto

// Represents a user sign in response
type SignInResponse struct {
	Status string `json:"status"`
	Role   string `json:"role"`
}
