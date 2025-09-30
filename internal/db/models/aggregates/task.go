package aggregates

import "time"

type Task struct {
	Id             int64
	ProjectId      int64
	Name           string
	Status         string
	Priority       string
	Description    string
	CreatedAt      time.Time
	EstTimeMinutes int64
	UserId         *int64
	FirstName      *string
	LastName       *string
	Title          *string
	Image          *string
}
