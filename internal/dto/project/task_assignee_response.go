package project

type TaskAssigneeResponse struct {
	Id        int64   `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Image     *string `json:"image"`
	Title     *string `json:"title"`
}
