package iam

type ConfirmRecoveryRequest struct {
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
