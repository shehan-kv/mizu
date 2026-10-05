package message

type MessageSenderResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Image     *string `json:"image"`
	Title     *string `json:"title"`
	Role      string  `json:"role"`
}
