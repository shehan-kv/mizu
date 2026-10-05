package task

type ListTaskParams struct {
	ActorID   string
	ProjectID string
	Keyword   *string
	Status    *string
	Priority  *string
	Limit     int
	Offset    int
}
