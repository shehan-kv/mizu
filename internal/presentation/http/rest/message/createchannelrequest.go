package message

type CreateChannelRequest struct {
	ProjectID *string  `json:"projectId"`
	Name      string   `json:"name"`
	MemberIDs []string `json:"memberIds"`
}
