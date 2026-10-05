package message

type CreateChannelParams struct {
	ActorID   string
	ProjectID *string
	Name      string
	MemberIDs []string
}
