package message

type ListProjectFilesParams struct {
	ActorID   string
	ProjectID string
	Keyword   *string
	Limit     int
	Offset    int
}
