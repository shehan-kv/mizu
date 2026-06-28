package iam

type RecoveryToken string

func NewRecoveryToken(token string) (RecoveryToken, error) {
	if token == "" {
		return "", ErrRecoveryTokenCannotBeEmpty
	}

	return RecoveryToken(token), nil
}

func (u RecoveryToken) String() string {
	return string(u)
}
