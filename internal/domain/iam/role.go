package iam

type Role string

const (
	RoleAdministrator Role = "administrator"
	RoleStaff         Role = "staff"
	RoleClient        Role = "client"
)

func NewRole(role string) (Role, error) {
	r := Role(role)

	switch r {
	case RoleAdministrator, RoleStaff, RoleClient:
		return r, nil
	default:
		return "", ErrUserInvalidRole
	}
}

func (r Role) String() string {
	return string(r)
}
