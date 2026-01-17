package params

type UserOnboard struct {
	FirstName  string
	LastName   string
	Title      *string
	Email      string
	Role       string
	IsActive   bool
	IsVerified bool
	Image      *string
	ActorID    int64
	Token      string
	Projects   []int64
}
