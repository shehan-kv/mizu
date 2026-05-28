package contract

type SignatoryStatus string

const (
	SignatoryStatusPending  SignatoryStatus = "pending"
	SignatoryStatusSigned   SignatoryStatus = "signed"
	SignatoryStatusRejected SignatoryStatus = "rejected"
)

func NewSignatoryStatus(status string) (SignatoryStatus, error) {
	s := SignatoryStatus(status)

	switch s {
	case SignatoryStatusPending, SignatoryStatusSigned, SignatoryStatusRejected:
		return s, nil
	default:
		return "", ErrContractInvalidSignatoryStatus
	}
}

func (s SignatoryStatus) String() string {
	return string(s)
}
