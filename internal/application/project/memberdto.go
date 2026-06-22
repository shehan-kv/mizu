package project

type MemberDTO struct {
	ID        string
	FirstName string
	LastName  string
	Title     *string
	HasImage  bool
	Role      string
}
