package contract

import "time"

type VersionResponse struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Contract  string    `json:"contract"`
}
