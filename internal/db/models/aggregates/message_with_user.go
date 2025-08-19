package aggregates

import "time"

type MessageWithUser struct {
	UserId    int64
	MessageId int64
	Message   string
	FirstName string
	LastName  string
	Type      string
	Role      string
	Title     string
	Image     *string
	CreatedAt time.Time
}
