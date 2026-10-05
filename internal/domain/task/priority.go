package task

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

func NewPriority(priority string) (Priority, error) {
	p := Priority(priority)

	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return p, nil
	default:
		return "", ErrTaskInvalidPriority
	}
}

func (p Priority) String() string {
	return string(p)
}
