package aggregates

import "time"

type Project struct {
	Id        int64
	Name      string
	CreatedAt time.Time
	Status    string
}
