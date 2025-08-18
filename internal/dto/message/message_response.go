package message

import "time"

// Represents a message response with sender information
type MessageResponse struct {
	UserId    int64
	MessageId int64
	Message   string
	FirstName string
	LastName  string
	Role      string
	Title     string
	Image     *string
	CreatedAt time.Time
}
