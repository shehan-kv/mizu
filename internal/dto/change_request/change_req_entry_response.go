package changerequest

import "time"

type ChangeReqEntryResponse struct {
	Id        int64                 `json:"id"`
	User      ChangeReqUserResponse `json:"user"`
	CreatedAt time.Time             `json:"createdAt"`
	Content   string                `json:"content"`
}
