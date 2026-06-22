package message

type MemberDTO struct {
	ID        string
	FirstName string
	LastName  string
	HasImage  bool
	Title     *string
	Role      string
}
