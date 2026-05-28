package task

type Stats struct {
	total     int
	completed int
}

func NewStats(total int, completed int) Stats {
	return Stats{total: total, completed: completed}
}

func (s Stats) Total() int {
	return s.total
}

func (s Stats) Completed() int {
	return s.completed
}
