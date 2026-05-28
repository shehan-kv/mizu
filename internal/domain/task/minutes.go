package task

type Minutes int

func NewMinutes(minutes int) (Minutes, error) {
	if minutes < 1 {
		return Minutes(0), ErrTaskMinutesMustBePositive
	}

	return Minutes(minutes), nil
}

func (m Minutes) Int() int {
	return int(m)
}
