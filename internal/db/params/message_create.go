package params

// Parameters to create a message
type MessageCreate struct {
	ChannelId int64
	UserId    int64
	Type      string
	Message   string
}
