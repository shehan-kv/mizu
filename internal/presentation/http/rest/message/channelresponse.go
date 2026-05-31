package message

import "time"

type ChannelResponse struct {
	ID        string    `json:"id"`
	ProjectID *string   `json:"projectId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
