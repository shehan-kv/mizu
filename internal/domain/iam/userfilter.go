package iam

type UserFilter struct {
	Keyword  *string
	Role     *Role
	IsActive *bool
}
