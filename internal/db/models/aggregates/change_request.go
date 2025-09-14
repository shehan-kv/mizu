package aggregates

import "time"

type ChangeRequest struct {
	Id               int64
	ProjectId        int64
	ProjectName      string
	ReqUserId        int64
	ReqUserFirstName string
	ReqUserLastName  string
	CreatedAt        time.Time
	Status           string
	Title            string
}
