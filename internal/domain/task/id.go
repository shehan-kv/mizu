package task

type TaskID string

func NewTaskID(id string) (TaskID, error) {
	if id == "" {
		return "", ErrTaskIDCannotBeEmpty
	}

	return TaskID(id), nil
}

func (t TaskID) String() string {
	return string(t)
}
