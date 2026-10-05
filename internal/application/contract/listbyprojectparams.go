package contract

type ListByProjectParams struct {
	ActorID   string
	ProjectID string
	Keyword   *string
	Status    *string
	Limit     int
	Offset    int
}
