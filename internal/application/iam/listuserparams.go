package iam

type ListUsersParams struct {
	ActorID    string
	Keyword    *string
	Role       *string
	IsActive   *bool
	IsVerified *bool
	Limit      int
	Offset     int
}
