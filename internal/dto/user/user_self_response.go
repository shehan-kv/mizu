package user

// Represents a user self response
type UserSelfResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Title     string `json:"title"`
	Image     string `json:"image"`
}
