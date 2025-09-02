package aggregates

import "time"

type ContractWithStats struct {
	Id                int64
	Name              string
	Status            string
	CreatedAt         time.Time
	Versions          int64
	Revisions         int64
	AcceptedRevisions int64
}
