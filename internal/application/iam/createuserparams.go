package iam

type CreateUserParams struct {
	ActorID    string
	FirstName  string
	LastName   string
	Email      string
	Title      *string
	Role       string
	IsActive   bool
	ProjectIDs []string
}
