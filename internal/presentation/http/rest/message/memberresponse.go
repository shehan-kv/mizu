package message

type MemberResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	HasImage  bool    `json:"hasImage"`
	Title     *string `json:"title"`
	Role      string  `json:"role"`
}
