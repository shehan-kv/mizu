package params

// Parameters to create a user verify request
type UserOnboardReqCreate struct {
	UserId  int64
	Token   string
	IsValid bool
}
