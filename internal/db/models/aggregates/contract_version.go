package aggregates

import "time"

type ContractVersion struct {
	Id        int64
	CreatedAt time.Time
	Status    string
	Version   string
	Contract  string
}
