package changerequest

type ChangeReqUserResponse struct {
	Id        int64   `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Title     *string `json:"title"`
	Role      *string `json:"role"`
	Image     *string `json:"image"`
}
