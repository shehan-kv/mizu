package iam

type SignInRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}
