package aggregates

import "time"

type ChangeRequestEntry struct {
	Id            int64
	UserId        int64
	UserFirstName string
	UserLastName  string
	UserTitle     *string
	UserImage     *string
	UserRole      string
	CreatedAt     time.Time
	Content       string
}
