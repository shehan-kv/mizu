package task

type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusInProgress Status = "in-progress"
	StatusCompleted  Status = "completed"
)

func NewStatus(status string) (Status, error) {
	s := Status(status)

	switch s {
	case StatusBacklog, StatusInProgress, StatusCompleted:
		return s, nil
	default:
		return "", ErrTaskInvalidStatus
	}
}

func (s Status) String() string {
	return string(s)
}
