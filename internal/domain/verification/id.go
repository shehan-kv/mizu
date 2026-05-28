package verification

type VerificationID string

func NewVerificationID(id string) (VerificationID, error) {

	if id == "" {
		return "", ErrVerificationIDCannotBeEmpty
	}

	return VerificationID(id), nil
}

func (v VerificationID) String() string {
	return string(v)
}
