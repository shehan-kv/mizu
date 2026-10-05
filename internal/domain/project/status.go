package project

type Status string

const (
	StatusStarted   Status = "started"
	StatusPaused    Status = "paused"
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
)

func NewStatus(status string) (Status, error) {
	s := Status(status)

	switch s {
	case StatusStarted, StatusPaused, StatusCancelled, StatusCompleted:
		return s, nil
	default:
		return "", ErrProjectInvalidStatus
	}
}

func (s Status) String() string {
	return string(s)
}
