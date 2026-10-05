package message

type ReplaceChannelMembersRequest struct {
	MemberIDs []string `json:"memberIds"`
}
