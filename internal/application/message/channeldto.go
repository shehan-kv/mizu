package message

import (
	"time"
)

type ChannelDTO struct {
	ID        string
	ProjectID *string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
