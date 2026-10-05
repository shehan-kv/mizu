package project

type MemberResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Title     *string `json:"title"`
	HasImage  bool    `json:"hasImage"`
	Role      string  `json:"role"`
}
