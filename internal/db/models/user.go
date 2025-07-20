package models

import "database/sql"

// Representation of user in database
type User struct {
	Id        int64
	FirstName string
	LastName  string
	Title     sql.NullString
	Email     string
	Role      int64
	Password  sql.NullString
	Image     sql.NullString
	CreatedAt sql.NullTime
	LastLogin sql.NullTime
	IsActive  bool
}
