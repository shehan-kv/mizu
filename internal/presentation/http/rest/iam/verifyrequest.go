package iam

type VerifyRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
