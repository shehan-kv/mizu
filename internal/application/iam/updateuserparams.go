package iam

type UpdateUserParams struct {
	ActorID   string
	UserID    string
	FirstName string
	LastName  string
	Email     string
	Title     *string
	Role      string
	Image     *ProfileImageParams
}
