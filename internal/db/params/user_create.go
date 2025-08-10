package params

import "database/sql"

// Parameters to create a user
type UserCreate struct {
	FirstName string
	LastName  string
	Title     sql.NullString
	Email     string
	Role      string
	IsActive  bool
	Image     sql.NullString
}
