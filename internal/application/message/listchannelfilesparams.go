package message

type ListChannelFilesParams struct {
	ActorID   string
	ChannelID string
	Keyword   *string
	Limit     int
	Offset    int
}
