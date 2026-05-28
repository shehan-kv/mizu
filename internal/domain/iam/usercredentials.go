package iam

type UserCredentials struct {
	user           *User
	hashedPassword string
}

func NewUserCredentials(user *User, hashedPw string) *UserCredentials {
	return &UserCredentials{
		user:           user,
		hashedPassword: hashedPw,
	}
}
