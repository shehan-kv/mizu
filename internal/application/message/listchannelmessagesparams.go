package message

type ListChannelMessagesParams struct {
	ActorID   string
	ChannelID string
	Limit     int
	Offset    int
}
