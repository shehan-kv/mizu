package billing

type Status string

var (
	StatusPaid      Status = "paid"
	StatusPending   Status = "pending"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
)

func NewStatus(status string) (Status, error) {
	s := Status(status)

	switch s {
	case StatusPaid, StatusPending, StatusAccepted, StatusRejected, StatusCancelled:
		return s, nil
	default:
		return "", ErrBillingInvalidStatus
	}
}

func (s Status) String() string {
	return string(s)
}
