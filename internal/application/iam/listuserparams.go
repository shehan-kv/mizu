package iam

type ListUsersParams struct {
	ActorID  string
	Keyword  *string
	Role     *string
	IsActive *bool
	Limit    int
	Offset   int
}
