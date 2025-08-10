package params

// Parameters to create a user onboard request
type UserOnboardRequestCreate struct {
	UserId  int64
	Token   string
	IsValid bool
}
