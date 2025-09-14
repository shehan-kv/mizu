package changerequest

import "time"

type ChangeReqResponse struct {
	Id          int64                     `json:"id"`
	Title       string                    `json:"title"`
	CreatedAt   time.Time                 `json:"createdAt"`
	RequestedBy *ChangeReqUserResponse    `json:"requestedBy"`
	Project     *ChangeReqProjectResponse `json:"project"`
	Status      string                    `json:"status"`
}
