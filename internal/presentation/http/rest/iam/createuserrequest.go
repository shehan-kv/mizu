package iam

type CreateUserRequest struct {
	FirstName  string   `json:"firstName"`
	LastName   string   `json:"lastName"`
	Email      string   `json:"email"`
	Title      *string  `json:"title"`
	Role       string   `json:"role"`
	IsActive   bool     `json:"isActive"`
	ProjectIDs []string `json:"projectIds"`
}
