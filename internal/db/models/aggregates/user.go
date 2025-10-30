package aggregates

import "time"

type User struct {
	Id        int64
	FirstName string
	LastName  string
	Title     *string
	Email     string
	Role      string
	Image     *string
	CreatedAt time.Time
	LastLogin *time.Time
	IsActive  bool
}
