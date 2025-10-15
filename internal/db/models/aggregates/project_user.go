package aggregates

type ProjectUser struct {
	Id        int64
	FirstName string
	LastName  string
	Title     *string
	Image     *string
	Role      string
}
