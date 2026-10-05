package task

type AssigneeDTO struct {
	ID        string
	FirstName string
	LastName  string
	Title     *string
	HasImage  bool
	Role      string
}
