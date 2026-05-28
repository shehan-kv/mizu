package iam

type UserID string

var SystemUserID = UserID("00000000-0000-7000-8000-000000000000")

func NewUserID(id string) (UserID, error) {
	if id == "" {
		return "", ErrUserIDCannotBeEmpty
	}

	return UserID(id), nil
}

func (u UserID) String() string {
	return string(u)
}
