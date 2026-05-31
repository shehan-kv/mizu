package task

type AssigneeResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Title     *string `json:"title"`
	Image     *string `json:"image"`
	Role      string  `json:"role"`
}
