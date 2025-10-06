package models

import "time"

type Channel struct {
	Id        int64
	ProjectId *int64
	Name      string
	CreatedAt time.Time
}
