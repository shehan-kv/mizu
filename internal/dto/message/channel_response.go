package message

import "time"

type ChannelResponse struct {
	Id        int64     `json:"id"`
	ProjectId *int64    `json:"projectId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
