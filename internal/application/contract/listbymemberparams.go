package contract

type ListByMemberParams struct {
	ActorID  string
	MemberID string
	Keyword  *string
	Status   *string
	Limit    int
	Offset   int
}
