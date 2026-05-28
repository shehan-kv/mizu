package contract

type Status string

const (
	StatusPending  Status = "pending"
	StatusRejected Status = "rejected"
	StatusSigned   Status = "signed"
)

func NewStatus(status string) (Status, error) {
	s := Status(status)

	switch s {
	case StatusPending, StatusRejected, StatusSigned:
		return s, nil
	default:
		return "", ErrContractInvalidStatus
	}
}

func (s Status) String() string {
	return string(s)
}
