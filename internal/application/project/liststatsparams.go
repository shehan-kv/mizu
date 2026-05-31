package project

type ListStatsParams struct {
	ActorID  string
	MemberID string
	Keyword  *string
	Status   *string
	Limit    int
	Offset   int
}
