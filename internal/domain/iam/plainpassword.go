package iam

type PlainPassword struct {
	value string
}

func NewPlainPassword(value string) (PlainPassword, error) {
	if len(value) < 8 {
		return PlainPassword{}, ErrUserPasswordTooShort
	}
	if len(value) > 72 { // bcrypt max length
		return PlainPassword{}, ErrUserPasswordTooLong
	}
	return PlainPassword{value: value}, nil
}

func (p PlainPassword) Value() string {
	return p.value
}

func (p PlainPassword) Equals(pw PlainPassword) bool {
	return p == pw
}
